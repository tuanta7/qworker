package handler

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/internal/usecase"
)

type SchedulerHandler struct {
	SchedulerUsecase *usecase.SchedulerUsecase
}

func NewSchedulerHandler(schedulerUsecase *usecase.SchedulerUsecase) *SchedulerHandler {
	return &SchedulerHandler{
		SchedulerUsecase: schedulerUsecase,
	}
}

func (h *SchedulerHandler) SendSyncMessage(ctx context.Context, task *asynq.Task) error {
	// Do something
	return nil
}
