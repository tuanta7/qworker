package handler

import (
	"context"
	"github.com/tuanta7/qworker/config"
	connectoruc "github.com/tuanta7/qworker/internal/connector"
	"github.com/tuanta7/qworker/internal/domain"
	scheduleruc "github.com/tuanta7/qworker/internal/scheduler"
)

type SchedulerHandler struct {
	cfg         *config.Config
	schedulerUC *scheduleruc.UseCase
	connectorUC *connectoruc.UseCase
}

func NewSchedulerHandler(
	cfg *config.Config,
	schedulerUC *scheduleruc.UseCase,
	connectorUC *connectoruc.UseCase,
) *SchedulerHandler {
	return &SchedulerHandler{
		cfg:         cfg,
		schedulerUC: schedulerUC,
		connectorUC: connectorUC,
	}
}

func (h *SchedulerHandler) HandleInsertConnector(ctx context.Context, connectorID uint64) error {
	connector, err := h.connectorUC.GetByID(ctx, connectorID)
	if err != nil || !connector.Enabled {
		return err
	}

	settings, err := connector.GetSyncSettings()
	if err != nil || !settings.IncSync {
		return err
	}

	message := &domain.QueueMessage{
		ConnectorID: connector.ConnectorID,
		TaskType:    domain.TaskTypeIncrementalSync,
		Queue:       config.LowQueue,
	}

	err = h.schedulerUC.CreateCronJob(settings.IncSyncPeriod, message)
	if err != nil {
		return err
	}

	h.schedulerUC.StartScheduler()
	return nil
}

func (h *SchedulerHandler) HandleUpdateConnector(ctx context.Context, connectorID uint64) error {
	connector, err := h.connectorUC.GetByID(ctx, connectorID)
	if err != nil {
		return err
	}

	if !connector.Enabled {
		h.schedulerUC.RemoveCronJob(connectorID)
		return nil
	}

	settings, err := connector.GetSyncSettings()
	if err != nil {
		return err
	}

	if !settings.IncSync {
		h.schedulerUC.RemoveCronJob(connectorID)
		return nil
	}

	currentPeriod, exists := h.schedulerUC.GetCronJobPeriod(connectorID)
	if exists {
		if currentPeriod == settings.IncSyncPeriod {
			return nil
		}
		h.schedulerUC.RemoveCronJob(connectorID)
	}

	message := &domain.QueueMessage{
		ConnectorID: connector.ConnectorID,
		TaskType:    domain.TaskTypeIncrementalSync,
		Queue:       config.LowQueue,
	}

	err = h.schedulerUC.CreateCronJob(settings.IncSyncPeriod, message)
	if err != nil {
		return err
	}

	h.schedulerUC.StartScheduler()
	return nil
}

func (h *SchedulerHandler) HandleDeleteConnector(ctx context.Context, connectorID uint64) {
	_ = h.schedulerUC.RemoveCronJob(connectorID)
}

func (h *SchedulerHandler) InitJobs(ctx context.Context) error {
	connectors, err := h.connectorUC.ListEnabled(ctx)
	if err != nil {
		return err
	}

	for _, connector := range connectors {
		message := &domain.QueueMessage{
			ConnectorID: connector.ConnectorID,
			TaskType:    domain.TaskTypeIncrementalSync,
			Queue:       config.LowQueue,
		}

		settings, err := connector.GetSyncSettings()
		if err != nil || !settings.IncSync {
			return err
		}

		err = h.schedulerUC.CreateCronJob(settings.IncSyncPeriod, message)
		if err != nil {
			return err
		}
	}

	h.schedulerUC.StartScheduler()
	return nil
}

func (h *SchedulerHandler) RemoveJobs() {
	h.schedulerUC.RemoveCronJobs()
}
