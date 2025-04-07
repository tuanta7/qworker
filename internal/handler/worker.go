package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/config"
	connectoruc "github.com/tuanta7/qworker/internal/connector"
	"github.com/tuanta7/qworker/internal/domain"
	"github.com/tuanta7/qworker/internal/usecase/worker"
	"github.com/tuanta7/qworker/pkg/logger"
	"time"
)

type WorkerHandler struct {
	workerUC    *workeruc.UseCase
	connectorUC *connectoruc.UseCase
	logger      *logger.ZapLogger
}

func NewWorkerHandler(workerUC *workeruc.UseCase, connectorUC *connectoruc.UseCase, zl *logger.ZapLogger) *WorkerHandler {
	return &WorkerHandler{
		workerUC:    workerUC,
		connectorUC: connectorUC,
		logger:      zl,
	}
}

func (h *WorkerHandler) HandleIncrementalSync(ctx context.Context, task *asynq.Task) error {
	message := &domain.QueueMessage{}
	err := json.Unmarshal(task.Payload(), message)
	if err != nil {
		return err
	}

	fullSyncTask, err := h.workerUC.GetTask(message.ConnectorID, config.QueueFullSync)
	if err != nil {
		return err
	}

	if fullSyncTask != nil {
		// terminate current task to run full sync task (w strict priority)
		return nil
	}

	err = h.workerUC.RunIncrementalSyncTask(ctx, message)
	if err != nil {
		return err
	}

	return nil
}

func (h *WorkerHandler) HandleFullSync(ctx context.Context, task *asynq.Task) error {
	message := &domain.QueueMessage{}
	err := json.Unmarshal(task.Payload(), message)
	if err != nil {
		return err
	}

	for {
		incSyncTask, err := h.workerUC.GetTask(message.ConnectorID, config.QueueIncrementalSync)
		if err != nil {
			if errors.Is(err, asynq.ErrTaskNotFound) || errors.Is(err, asynq.ErrQueueNotFound) {
				break
			}
			return err
		}

		if incSyncTask != nil && incSyncTask.State == asynq.TaskStateActive {
			time.Sleep(10 * time.Second)
			continue
		}
		break
	}

	err = h.workerUC.RunFullSyncTask(ctx, message)
	if err != nil {
		return err
	}
	return nil
}
