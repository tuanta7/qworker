package handler

import (
	"context"
	"encoding/json"
	"github.com/hibiken/asynq"
	connectoruc "github.com/tuanta7/qworker/internal/connector"
	"github.com/tuanta7/qworker/internal/domain"
	workeruc "github.com/tuanta7/qworker/internal/worker"
	"github.com/tuanta7/qworker/pkg/logger"
	"github.com/tuanta7/qworker/pkg/utils"
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

func (h *WorkerHandler) HandleTask(ctx context.Context, task *asynq.Task) error {
	message := &domain.QueueMessage{}
	err := json.Unmarshal(task.Payload(), message)
	if err != nil {
		// Shut down task and remove from queue
		return nil
	}

	// Prevent failed task staying in queue
	defer h.workerUC.CleanTask(message.Queue, message.ConnectorID)

	currentTask, err := h.workerUC.IsConnectorRunning(message.ConnectorID)
	if err != nil {
		return err
	}

	if currentTask != nil {
		if domain.QueuePriority[currentTask.Queue] >= domain.QueuePriority[message.Queue] {
			return utils.ErrTaskConflict
		}

		err = h.workerUC.TerminateTask(message.Queue, message.ConnectorID)
		if err != nil {
			return err
		}
	}

	err = h.workerUC.RunTask(ctx, message)
	if err != nil {
		return err
	}

	return nil
}
