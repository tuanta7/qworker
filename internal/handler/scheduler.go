package handler

import (
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/tuanta7/qworker/internal/usecase"
	"github.com/tuanta7/qworker/pkg/logger"
)

type SchedulerHandler struct {
	jobMutex sync.Mutex
	jobMap   map[uint64]cron.EntryID

	schedulerUC *usecase.SchedulerUsecase
	logger      *logger.ZapLogger
	cronClient  *cron.Cron
}

func NewSchedulerHandler(schedulerUC *usecase.SchedulerUsecase, logger *logger.ZapLogger, cronClient *cron.Cron) *SchedulerHandler {
	return &SchedulerHandler{
		jobMap:      make(map[uint64]cron.EntryID),
		schedulerUC: schedulerUC,
		logger:      logger,
		cronClient:  cronClient,
	}
}

func (h *SchedulerHandler) SendSyncMessage(connectorID uint64, interval time.Duration) error {
	jobID, err := h.cronClient.AddFunc(
		fmt.Sprintf("@every %s", interval.String()),
		h.schedulerUC.SendSyncMessage(connectorID),
	)
	if err != nil {
		return err
	}
	h.cronClient.Start()

	h.jobMutex.Lock()
	h.jobMap[connectorID] = jobID
	h.jobMutex.Unlock()

	return nil
}

func (h *SchedulerHandler) TerminateJob(connectorID uint64) error {
	h.jobMutex.Lock()
	defer h.jobMutex.Unlock()

	if jobID, ok := h.jobMap[connectorID]; ok {
		h.cronClient.Remove(jobID)
		delete(h.jobMap, connectorID)
		return nil
	}

	return fmt.Errorf("job not found: %d", connectorID)
}
