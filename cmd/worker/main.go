package main

import (
	"log"

	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/config"
	"github.com/tuanta7/qworker/internal/domain"
	"github.com/tuanta7/qworker/internal/handler"
	pgrepo "github.com/tuanta7/qworker/internal/repository/postgres"
	workeruc "github.com/tuanta7/qworker/internal/worker"
	"github.com/tuanta7/qworker/pkg/db"
	"github.com/tuanta7/qworker/pkg/logger"
)

func main() {
	cfg := config.NewConfig()
	zapLogger := logger.MustNewLogger(cfg.Logger.Level)

	pgClient, err := db.NewPostgresClient(cfg)
	if err != nil {
		log.Fatalf("db.NewPostgresClient: %v", err)
	}
	defer pgClient.Close()

	connectorRepository := pgrepo.NewConnectorRepository(pgClient)
	workerUseCase := workeruc.NewUseCase(connectorRepository, zapLogger)

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
				domain.IncrementalSyncJobQueue: 1,
				domain.FullSyncJobQueue:        5,
			},
		})

	mux := NewRouter(cfg, zapLogger, workerUseCase)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("Asynq server stopped: %v", err)
	}
}

func NewRouter(cfg *config.Config, zl *logger.ZapLogger, workerUC *workeruc.UseCase) *asynq.ServeMux {
	workerHandler := handler.NewWorkerHandler(workerUC, zl)

	mux := asynq.NewServeMux()
	mux.HandleFunc(domain.IncrementalSyncJobQueue, workerHandler.HandleUserSync)
	mux.HandleFunc(domain.FullSyncJobQueue, workerHandler.HandleUserSync)
	return mux
}
