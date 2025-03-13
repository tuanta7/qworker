package handler

import (
	"context"
	"encoding/json"
	connectoruc "github.com/tuanta7/qworker/internal/connector"
	"github.com/tuanta7/qworker/internal/domain"

	"github.com/hibiken/asynq"
	workeruc "github.com/tuanta7/qworker/internal/worker"
	"github.com/tuanta7/qworker/pkg/logger"
	"go.uber.org/zap"
)

type WorkerHandler struct {
	workerUC    *workeruc.UseCase
	connectorUC *connectoruc.UseCase
	logger      *logger.ZapLogger
}

func NewWorkerHandler(workerUC *workeruc.UseCase, connectorUC *connectoruc.UseCase, logger *logger.ZapLogger) *WorkerHandler {
	return &WorkerHandler{
		workerUC:    workerUC,
		connectorUC: connectorUC,
		logger:      logger,
	}
}

func (h *WorkerHandler) HandleUserIncrementalSync(ctx context.Context, task *asynq.Task) error {
	message := domain.QueueMessage{}
	err := json.Unmarshal(task.Payload(), &message)
	if err != nil {
		return err
	}

	err = h.workerUC.RunTask(ctx, message)
	if err != nil {
		return err
	}

	h.logger.Info("WorkerHandler - HandleUserSync", zap.Any("task", message))
	return nil
}

func (h *WorkerHandler) HandleUserFullSync(ctx context.Context, task *asynq.Task) error {
	h.logger.Info("WorkerHandler - HandleUserFullSync", zap.Any("task", task.Payload()))
	return nil
}

func (h *WorkerHandler) HandleTerminateSync(ctx context.Context, task *asynq.Task) error {
	h.logger.Info("WorkerHandler - HandleUserSync", zap.Any("task", task.Payload()))
	return nil
}
