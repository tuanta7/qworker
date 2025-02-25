package main

import (
	"log"

	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/config"
	pgrepo "github.com/tuanta7/qworker/internal/repository/postgres"
	"github.com/tuanta7/qworker/internal/usecase"
	"github.com/tuanta7/qworker/pkg/db"
)

func main() {
	cfg := config.NewConfig()

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

	schedulerRepository := pgrepo.NewSchedulerRepository(pgClient)
	schedulerUsecase := usecase.NewSchedulerUsecase(schedulerRepository)

	// This should be done in a cron job, with interval
	task := asynq.NewTask("user:sync", schedulerUsecase.NewSyncMessage())
	if _, err := asynqClient.Enqueue(task); err != nil {
		log.Fatalf("Enqueue task: %v", err)
	}
}
