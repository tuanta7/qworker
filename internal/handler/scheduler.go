package handler

import (
	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/internal/usecase"
)

type SchedulerHandler struct {
	asynqClient *asynq.Client
	schedulerUC *usecase.SchedulerUsecase
}

func NewSchedulerHandler(asynqClient *asynq.Client, schedulerUC *usecase.SchedulerUsecase) *SchedulerHandler {
	return &SchedulerHandler{
		asynqClient: asynqClient,
		schedulerUC: schedulerUC,
	}
}

func (h *SchedulerHandler) CreateNewJob() {
	// This should be done in a cron job, with interval
	task := asynq.NewTask("user:sync", h.schedulerUC.NewSyncMessage())
	if _, err := h.asynqClient.Enqueue(task); err != nil {
		log.Fatalf("Enqueue task: %v", err)
	}
}