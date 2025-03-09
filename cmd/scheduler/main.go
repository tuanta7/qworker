package main

import (
	"context"
	"encoding/json"
	redisrepo "github.com/tuanta7/qworker/internal/repository/redis"
	"log"
	"strings"

	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/config"
	"github.com/tuanta7/qworker/internal/handler"
	pgrepo "github.com/tuanta7/qworker/internal/repository/postgres"
	scheduleruc "github.com/tuanta7/qworker/internal/scheduler"
	"github.com/tuanta7/qworker/pkg/db"
	"github.com/tuanta7/qworker/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	cfg := config.NewConfig()
	zapLogger := logger.MustNewLogger(cfg.Logger.Level)

	pgClient, err := db.NewPostgresClient(cfg)
	if err != nil {
		log.Fatalf("db.NewPostgresClient: %v", err)
	}
	defer pgClient.Close()

	asynqClient := asynq.NewClient(asynq.RedisFailoverClientOpt{
		MasterName:    cfg.Redis.MasterName,
		SentinelAddrs: cfg.Redis.Sentinels,
		Password:      cfg.Redis.Password,
		DB:            cfg.Redis.Database,
	})
	defer asynqClient.Close()

	connectorRepository := pgrepo.NewConnectorRepository(pgClient)
	jobRepository := redisrepo.NewJobRepository(asynqClient)
	schedulerUsecase := scheduleruc.NewUseCase(connectorRepository, jobRepository, zapLogger)
	schedulerHandler := handler.NewSchedulerHandler(schedulerUsecase, zapLogger)

	err = schedulerHandler.InitScheduledJobs()
	if err != nil {
		log.Fatalf("schedulerHandler.InitScheduledJobs(): %v", err)
	}

	// Listen for database changes
	go listen(ctx, pgClient, zapLogger)

	// Block the main goroutine with an empty select statement
	// to allow other goroutines to run
	select {}
}

func listen(ctx context.Context, pgClient *db.PostgresClient, zapLogger *logger.ZapLogger) {
	conn, err := pgClient.Pool.Acquire(ctx)
	if err != nil {
		zapLogger.Error(
			"failed to acquire database connection to listen for notifications",
			zap.Error(err),
		)
		return
	}

	_, err = conn.Exec(ctx, "LISTEN connectors_changes")
	if err != nil {
		zapLogger.Error(
			"failed to listen for notifications",
			zap.Error(err),
		)
		return
	}
	defer func() {
		_, _ = conn.Exec(ctx, "UNLISTEN connectors_changes")
		conn.Release()
	}()

	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			zapLogger.Error("conn.Conn().WaitForNotification", zap.Error(err))
			return
		}

		zapLogger.Info("received notification", zap.Any("payload", notification.Payload))

		if notification.Channel != "connectors_changes" {
			continue
		}

		message := &struct {
			Table       string `json:"table"`
			Action      string `json:"action"`
			ConnectorID uint64 `json:"connector_id"`
		}{}

		err = json.Unmarshal([]byte(notification.Payload), message)
		if err != nil {
			zapLogger.Error("failed to unmarshal notification", zap.Error(err))
		}

		switch strings.ToLower(message.Action) {
		case "insert":
			zapLogger.Info("Inserted")
		case "update":
			// Handle update event
		case "delete":
			zapLogger.Info("Deleted")
		default:
			zapLogger.Warn("unknown notification payload")
		}
	}
}
