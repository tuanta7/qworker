package main

import (
	"context"
	"log"

	"github.com/hibiken/asynq"
	"github.com/robfig/cron/v3"
	"github.com/tuanta7/qworker/config"
	"github.com/tuanta7/qworker/internal/handler"
	pgrepo "github.com/tuanta7/qworker/internal/repository/postgres"
	scheduleruc "github.com/tuanta7/qworker/internal/usecase/scheduler"
	"github.com/tuanta7/qworker/pkg/db"
	"github.com/tuanta7/qworker/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfig()
	zapLogger := logger.MustNewLogger(cfg.Logger.Level)

	ctx := context.Background()

	pgClient, err := db.NewPostgresClient(cfg)
	if err != nil {
		log.Fatalf("Postgres: %v", err)
	}
	defer pgClient.Close()

	asynqClient := asynq.NewClient(asynq.RedisFailoverClientOpt{
		MasterName:    cfg.Redis.MasterName,
		SentinelAddrs: cfg.Redis.Sentinels,
		Password:      cfg.Redis.Password,
		DB:            cfg.Redis.Database,
	})
	defer asynqClient.Close()

	cronClient := cron.New(cron.WithSeconds())

	schedulerRepository := pgrepo.NewSchedulerRepository(pgClient)
	schedulerUsecase := scheduleruc.NewSchedulerUsecase(schedulerRepository, zapLogger, asynqClient)
	schedulerHandler := handler.NewSchedulerHandler(schedulerUsecase, zapLogger, cronClient)

	schedulerHandler.InitScheduledJobs()
	go Listen(ctx, pgClient, zapLogger)

	// Block the main goroutine with an empty select statement
	// to allow other goroutines to run
	select {}
}

func Listen(ctx context.Context, pgClient *db.PostgresClient, zapLogger *logger.ZapLogger) {
	// Get a connection from the pool and never release it
	conn, err := pgClient.Pool.Acquire(ctx)
	if err != nil {
		zapLogger.Error(
			"failed to acquire database connection to listen for notifications",
			zap.Error(err),
		)
		return
	}

	// Listen for notifications on the "connectors_changes" channel
	_, err = conn.Exec(ctx, "LISTEN connectors_changes")
	if err != nil {
		zapLogger.Error(
			"failed to listen for notifications",
			zap.Error(err),
		)
		return
	}
	defer func() {
		conn.Exec(ctx, "UNLISTEN connectors_changes")
		conn.Release()
	}()

	// Wait for notifications in an infinite loop
	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			zapLogger.Error("conn.Conn().WaitForNotification", zap.Error(err))
			return
		}

		if notification.Channel != "connectors_changes" {
			continue
		}

		switch notification.Payload {
		case "insert":
			// Handle insert event
		case "update":
			// Handle update event
		case "delete":
			// Handle delete event
		default:
			zapLogger.Warn("unknown notification payload")
		}

		zapLogger.Info("received notification", zap.Any("payload", notification.Payload))
	}
}
