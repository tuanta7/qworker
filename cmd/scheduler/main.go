package main

import (
	"context"
	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/config"
	connectoruc "github.com/tuanta7/qworker/internal/connector"
	"github.com/tuanta7/qworker/internal/handler"
	pgrepo "github.com/tuanta7/qworker/internal/repository/postgres"
	redisrepo "github.com/tuanta7/qworker/internal/repository/redis"
	"github.com/tuanta7/qworker/internal/usecase/scheduler"
	"github.com/tuanta7/qworker/pkg/db"
	"github.com/tuanta7/qworker/pkg/logger"
	"log"
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

	err := schedulerHandler.Init(context.Background())
	if err != nil {
		log.Fatalf("schedulerHandler.InitScheduledJobs(): %v", err)
	}
	defer schedulerHandler.Clear()

	s := NewScheduler(pgClient, zapLogger)
	s.RegisterHandler("insert", schedulerHandler.HandleInsertConnector)
	s.RegisterHandler("update", schedulerHandler.HandleUpdateConnector)
	s.RegisterHandler("delete", schedulerHandler.HandleDeleteConnector)
	s.Listen(context.Background(), "connectors_changes")
}
