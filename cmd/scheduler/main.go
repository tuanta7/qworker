package main

import (
	"context"
	"encoding/json"
	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/config"
	connectoruc "github.com/tuanta7/qworker/internal/connector"
	"github.com/tuanta7/qworker/internal/handler"
	pgrepo "github.com/tuanta7/qworker/internal/repository/postgres"
	redisrepo "github.com/tuanta7/qworker/internal/repository/redis"
	scheduleruc "github.com/tuanta7/qworker/internal/scheduler"
	"github.com/tuanta7/qworker/pkg/db"
	"github.com/tuanta7/qworker/pkg/logger"
	"go.uber.org/zap"
	"log"
	"strings"
)

func main() {
	cfg := config.NewConfig()
	zapLogger := logger.MustNewLogger(cfg.Logger.Level)

	pgClient := db.MustNewPostgresClient(cfg)
	defer pgClient.Close()

	redisClient := db.MustNewRedisSentinelClient(cfg)
	defer redisClient.Close()

	asynqClient := asynq.NewClientFromRedisClient(redisClient)
	defer asynqClient.Close()

	asynqInspector := asynq.NewInspectorFromRedisClient(redisClient)
	defer asynqInspector.Close()

	taskRepository := redisrepo.NewTaskRepository(redisClient)
	connectorRepository := pgrepo.NewConnectorRepository(pgClient)
	connectorUsecase := connectoruc.NewUseCase(connectorRepository, zapLogger)
	schedulerUsecase := scheduleruc.NewUseCase(asynqClient, asynqInspector, taskRepository, zapLogger)
	schedulerHandler := handler.NewSchedulerHandler(cfg, schedulerUsecase, connectorUsecase)

	err := schedulerHandler.InitJobs(context.Background())
	if err != nil {
		log.Fatalf("schedulerHandler.InitScheduledJobs(): %v", err)
	}
	defer schedulerHandler.RemoveJobs()

	// Block the main goroutine and listen
	listen(pgClient, zapLogger, schedulerHandler)
}

func listen(pgClient *db.PostgresClient, zapLogger *logger.ZapLogger, schedulerHandler *handler.SchedulerHandler) {
	ctx := context.Background()

	conn, err := pgClient.Pool.Acquire(ctx)
	if err != nil {
		zapLogger.Error("failed to acquire database connection", zap.Error(err))
		return
	}

	_, err = conn.Exec(ctx, "LISTEN connectors_changes")
	if err != nil {
		zapLogger.Error("failed to listen for notifications", zap.Error(err))
		return
	}
	defer func() {
		_, _ = conn.Exec(ctx, "UNLISTEN connectors_changes")
		conn.Release()
	}()

	notifyChan := make(chan string, 10)
	go processNotifications(notifyChan, schedulerHandler, zapLogger)

	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			zapLogger.Error("conn.Conn().WaitForNotification", zap.Error(err))
			continue
		}

		if notification.Channel == "connectors_changes" {
			notifyChan <- notification.Payload
		}
	}
}

func processNotifications(notifyChan <-chan string, schedulerHandler *handler.SchedulerHandler, zapLogger *logger.ZapLogger) {
	for n := range notifyChan {
		message := db.NotifyMessage{}
		err := json.Unmarshal([]byte(n), &message)
		if err != nil {
			zapLogger.Error("failed to unmarshal notification", zap.Error(err))
		}

		schedulerCtx := context.Background()
		switch strings.ToLower(message.Action) {
		case "insert":
			zapLogger.Info("connector inserted", zap.Any("message", message))
			err = schedulerHandler.HandleInsertConnector(schedulerCtx, message.ID)
		case "update":
			zapLogger.Info("connector updated", zap.Any("message", message))
			err = schedulerHandler.HandleUpdateConnector(schedulerCtx, message.ID)
		case "delete":
			zapLogger.Info("connector deleted", zap.Any("message", message))
			schedulerHandler.HandleDeleteConnector(schedulerCtx, message.ID)
		}

		if err != nil {
			zapLogger.Error("failed to handle trigger action", zap.Error(err))
		}
	}
}
