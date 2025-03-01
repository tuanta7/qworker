package main

import (
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/robfig/cron/v3"
	"github.com/tuanta7/qworker/config"
	"github.com/tuanta7/qworker/internal/handler"
	pgrepo "github.com/tuanta7/qworker/internal/repository/postgres"
	"github.com/tuanta7/qworker/internal/usecase"
	"github.com/tuanta7/qworker/pkg/db"
	"github.com/tuanta7/qworker/pkg/logger"
)

func main() {
	cfg := config.NewConfig()
	zapLogger := logger.MustNewLogger(cfg.Logger.Level)

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
	schedulerUsecase := usecase.NewSchedulerUsecase(schedulerRepository, zapLogger, asynqClient)
	schedulerHandler := handler.NewSchedulerHandler(schedulerUsecase, zapLogger, cronClient)

	go func() {
		// Wait for database trigger
	}()
	err = schedulerHandler.SendSyncMessage(1, 60*time.Second)
	fmt.Println(err)

	select {}
}
