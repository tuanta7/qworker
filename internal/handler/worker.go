package handler

import (
	"context"

	"github.com/hibiken/asynq"
	workeruc "github.com/tuanta7/qworker/internal/worker"
	"github.com/tuanta7/qworker/pkg/logger"
	"go.uber.org/zap"
)

type WorkerHandler struct {
	workerUC *workeruc.UseCase
	logger   *logger.ZapLogger
}

func NewWorkerHandler(workerUC *workeruc.UseCase, logger *logger.ZapLogger) *WorkerHandler {
	return &WorkerHandler{
		workerUC: workerUC,
		logger:   logger,
	}
}

func (h *WorkerHandler) HandleTerminateSync(ctx context.Context, task *asynq.Task) error {
	h.logger.Info("WorkerHandler - HandleUserSync", zap.Any("task", task))
	return nil
}

func (h *WorkerHandler) HandleUserIncrementalSync(ctx context.Context, task *asynq.Task) error {
	h.logger.Info("WorkerHandler - HandleUserSync", zap.Any("task", task))
	return nil
}

func (h *WorkerHandler) HandleUserFullSync(ctx context.Context, task *asynq.Task) error {
	h.logger.Info("WorkerHandler - HandleUserFullSync", zap.Any("task", task))
	// Check if any Full/Incremental Sync Job is running
	// Skip if there is a full sync job still running, override if Incr. Sync
	return nil
}
