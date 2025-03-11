package handler

import (
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/config"
	"github.com/tuanta7/qworker/internal/domain"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	scheduleruc "github.com/tuanta7/qworker/internal/scheduler"
	"github.com/tuanta7/qworker/pkg/logger"
)

type SchedulerHandler struct {
	lock sync.Mutex
	jobs map[uint64]cron.EntryID

	schedulerUC *scheduleruc.UseCase
	logger      *logger.ZapLogger
	scheduler   *cron.Cron
}

func NewSchedulerHandler(schedulerUC *scheduleruc.UseCase, logger *logger.ZapLogger) *SchedulerHandler {
	return &SchedulerHandler{
		jobs:        make(map[uint64]cron.EntryID),
		schedulerUC: schedulerUC,
		logger:      logger,
		scheduler:   cron.New(cron.WithSeconds()),
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
	h.lock.Lock()
	defer h.lock.Unlock()

	message := domain.Message{
		ConnectorID: connectorID,
		JobType:     domain.JobTypeIncrementalSync,
	}

	payload, err := json.Marshal(message)
	if err != nil {

	}

	jobID, err := h.scheduler.AddFunc(
		fmt.Sprintf("@every %s", interval.String()),
		h.schedulerUC.EnqueueTask(asynq.NewTask(config.IncrementalSyncQueue, payload)),
	)
	if err != nil {
		return err
	}

	h.jobs[connectorID] = jobID
	h.scheduler.Start()
	return nil
}

func (h *SchedulerHandler) TerminateJob(connectorID uint64) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	if jobID, ok := h.jobs[connectorID]; ok {
		h.scheduler.Remove(jobID)
		delete(h.jobs, connectorID)
		return nil
	}

	return fmt.Errorf("job not found: %d", connectorID)
}
