package handler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/internal/usecase"
)

var jobMutex sync.Mutex

type SchedulerHandler struct {
	asynqClient *asynq.Client
	schedulerUC *usecase.SchedulerUsecase
	jobMap      map[uint64]context.CancelFunc
}

func NewSchedulerHandler(asynqClient *asynq.Client, schedulerUC *usecase.SchedulerUsecase) *SchedulerHandler {
	return &SchedulerHandler{
		asynqClient: asynqClient,
		schedulerUC: schedulerUC,
	}
}

func (h *SchedulerHandler) CreateNewJob(connectorID uint64, interval time.Duration) {
	ctx, cancel := context.WithCancel(context.Background())

	jobMutex.Lock()
	h.jobMap[connectorID] = cancel
	jobMutex.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		select {
		case <-ctx.Done():
			fmt.Printf("Stopped job: %d\n", connectorID)
			return
		case <-ticker.C:
			task := asynq.NewTask("user:sync", h.schedulerUC.NewSyncMessage())
			if _, err := h.asynqClient.Enqueue(task); err != nil {
				log.Printf("Enqueue task: %v", err)
			}
		}
	}()
}

func (h *SchedulerHandler) TerminateJob(connectorID uint64) {

}
