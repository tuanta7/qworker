package main

import (
	connectoruc "github.com/tuanta7/qworker/internal/connector"
	"log"

	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/config"
	pgrepo "github.com/tuanta7/qworker/internal/repository/postgres"
	workeruc "github.com/tuanta7/qworker/internal/worker"
	"github.com/tuanta7/qworker/pkg/db"
	"github.com/tuanta7/qworker/pkg/logger"
)

func main() {
	cfg := config.NewConfig()
	zapLogger := logger.MustNewLogger(cfg.Logger.Level)

	pgClient, err := db.NewPostgresClient(cfg, db.WithMaxConns(10))
	if err != nil {
		log.Fatalf("db.NewPostgresClient: %v", err)
	}
	defer pgClient.Close()

	connectorRepository := pgrepo.NewConnectorRepository(pgClient)
	connectorUsecase := connectoruc.NewUseCase(connectorRepository, zapLogger)
	workerUsecase := workeruc.NewUseCase(connectorRepository, zapLogger)

	srv := asynq.NewServer(
		asynq.RedisFailoverClientOpt{
			MasterName:    cfg.Redis.MasterName,
			SentinelAddrs: cfg.Redis.Sentinels,
			Password:      cfg.Redis.Password,
			DB:            cfg.Redis.Database,
		},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				config.TerminateQueue:       6,
				config.FullSyncQueue:        3,
				config.IncrementalSyncQueue: 1,
			},
		})

	mux := NewRouter(cfg, zapLogger, workerUsecase, connectorUsecase)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("asynq server stopped: %v", err)
	}
}
