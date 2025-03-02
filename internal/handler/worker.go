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

func (h *WorkerHandler) HandleUserSync(ctx context.Context, task *asynq.Task) error {
	h.logger.Info("WorkerHandler - HandleUserSync", zap.Any("task", task))
	return nil
}
