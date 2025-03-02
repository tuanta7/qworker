package handler

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/pkg/logger"
)

type WorkerHandler struct {
	logger *logger.ZapLogger
}

func NewWorkerHandler(logger *logger.ZapLogger) *WorkerHandler {
	return &WorkerHandler{
		logger: logger,
	}
}

func (h *WorkerHandler) HandleUserSync(ctx context.Context, task *asynq.Task) error {
	h.logger.Info("WorkerHandler - HandleUserSync")
	return nil
}
