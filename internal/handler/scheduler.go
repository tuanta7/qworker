package handler

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/tuanta7/qworker/internal/usecase"
)

var jobMutex sync.Mutex

type SchedulerHandler struct {
	schedulerUC *usecase.SchedulerUsecase
	jobMap      map[uint64]cron.EntryID
	cronClient  *cron.Cron
}

func NewSchedulerHandler(schedulerUC *usecase.SchedulerUsecase, cronClient *cron.Cron) *SchedulerHandler {
	return &SchedulerHandler{
		schedulerUC: schedulerUC,
		cronClient:  cronClient,
		jobMap:      make(map[uint64]cron.EntryID),
	}
}

func (h *SchedulerHandler) SendSyncMessage(connectorID uint64, interval time.Duration) error {
	job := h.schedulerUC.SendSyncMessage(connectorID)
	jobID, err := h.cronClient.AddFunc(fmt.Sprintf("@every %s", interval), job)
	if err != nil {
		log.Printf("Add cron job: %v", err)
		return err
	}

	jobMutex.Lock()
	h.jobMap[connectorID] = jobID
	jobMutex.Unlock()

	return nil
}

func (h *SchedulerHandler) TerminateJob(connectorID uint64) error {
	jobMutex.Lock()
	defer jobMutex.Unlock()

	if jobID, ok := h.jobMap[connectorID]; ok {
		h.cronClient.Remove(jobID)
		delete(h.jobMap, connectorID)
		return nil
	}

	return fmt.Errorf("job not found: %d", connectorID)
}
