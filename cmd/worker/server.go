package main

import (
	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/config"
	connectoruc "github.com/tuanta7/qworker/internal/connector"
	"github.com/tuanta7/qworker/internal/domain"
	"github.com/tuanta7/qworker/internal/handler"
	workeruc "github.com/tuanta7/qworker/internal/worker"
	"github.com/tuanta7/qworker/pkg/logger"
)

func NewRouter(
	cfg *config.Config,
	zl *logger.ZapLogger,
	workerUC *workeruc.UseCase,
	connectorUC *connectoruc.UseCase,
) *asynq.ServeMux {
	workerHandler := handler.NewWorkerHandler(workerUC, connectorUC, zl)

	mux := asynq.NewServeMux()
	mux.HandleFunc(domain.TaskTypeIncrementalSync, workerHandler.HandleTask)
	mux.HandleFunc(domain.TaskTypeFullSync, workerHandler.HandleTask)
	mux.HandleFunc(domain.TaskTypeTerminate, workerHandler.HandleTask)

	return mux
}
