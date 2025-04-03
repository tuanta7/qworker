package main

import (
	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/config"
	connectoruc "github.com/tuanta7/qworker/internal/connector"
	"github.com/tuanta7/qworker/internal/handler"
	"github.com/tuanta7/qworker/internal/usecase/worker"
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
	mux.HandleFunc(config.QueueTask[config.QueueIncrementalSync], workerHandler.HandleIncrementalSync)
	mux.HandleFunc(config.QueueTask[config.QueueFullSync], workerHandler.HandleFullSync)

	return mux
}
