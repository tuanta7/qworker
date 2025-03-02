package handler

import (
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	scheduleruc "github.com/tuanta7/qworker/internal/scheduler"
	"github.com/tuanta7/qworker/pkg/logger"
)

type SchedulerHandler struct {
	jobMutex sync.Mutex
	jobMap   map[uint64]cron.EntryID

	schedulerUC *scheduleruc.UseCase
	logger      *logger.ZapLogger
	cronClient  *cron.Cron
}

func NewSchedulerHandler(schedulerUC *scheduleruc.UseCase, logger *logger.ZapLogger, cronClient *cron.Cron) *SchedulerHandler {
	return &SchedulerHandler{
		jobMap:      make(map[uint64]cron.EntryID),
		schedulerUC: schedulerUC,
		logger:      logger,
		cronClient:  cronClient,
	}
}

func (h *SchedulerHandler) InitScheduledJobs() error {
	// Load connectors from database
	// connectors, err := h.schedulerUC.GetConnectors()
	// if err != nil {
	// 	return err
	// }

	// for _, connector := range connectors {
	// 	err := h.SendSyncMessage()
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (h *SchedulerHandler) SendSyncMessage(connectorID uint64, interval time.Duration) error {
	h.jobMutex.Lock()
	defer h.jobMutex.Unlock()

	jobID, err := h.cronClient.AddFunc(
		fmt.Sprintf("@every %s", interval.String()),
		h.schedulerUC.SendSyncJob(connectorID),
	)
	if err != nil {
		return err
	}

	h.jobMap[connectorID] = jobID
	h.cronClient.Start()
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
