package handler

import (
	"context"
	"encoding/json"
	connectoruc "github.com/tuanta7/qworker/internal/connector"
	"github.com/tuanta7/qworker/internal/domain"

	"github.com/hibiken/asynq"
	workeruc "github.com/tuanta7/qworker/internal/worker"
	"github.com/tuanta7/qworker/pkg/logger"
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

func (h *WorkerHandler) HandleTerminateSync(ctx context.Context, task *asynq.Task) error {
	message := &domain.QueueMessage{}
	err := json.Unmarshal(task.Payload(), message)
	if err != nil {
		return err
	}

	h.workerUC.TerminateTask(message.ConnectorID)
	return nil
}

func (h *WorkerHandler) HandleUserIncrementalSync(ctx context.Context, task *asynq.Task) error {
	return h.handleUserSync(ctx, task)
}

func (h *WorkerHandler) HandleUserFullSync(ctx context.Context, task *asynq.Task) error {
	return h.handleUserSync(ctx, task)
}

func (h *WorkerHandler) handleUserSync(ctx context.Context, task *asynq.Task) error {
	message := &domain.QueueMessage{}
	err := json.Unmarshal(task.Payload(), message)
	if err != nil {
		return err
	}

	err = h.workerUC.RunTask(ctx, message)
	if err != nil {
		return err
	}
	return nil
}
