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

func (h *SchedulerHandler) ClearCronJobs() {
	h.schedulerUC.ClearCronJobs()
}

func (h *SchedulerHandler) HandleDeleteConnector(ctx context.Context, connectorID uint64) {
	_ = h.schedulerUC.RemoveCronJob(connectorID)
}

func (h *SchedulerHandler) InitCronJobs(ctx context.Context) error {
	connectors, err := h.connectorUC.ListEnabled(ctx)
	if err != nil {
		return err
	}

	for _, connector := range connectors {
		settings, err := connector.GetSyncSettings()
		if err != nil || !settings.IncSync {
			continue
		}

		message := &domain.QueueMessage{
			ConnectorID: connector.ConnectorID,
			TaskType:    domain.TaskTypeIncrementalSync,
			Queue:       domain.LowQueue,
		}

		err = h.schedulerUC.CreateCronJob(settings.IncSyncPeriod, message)
		if err != nil {
			return err
		}
	}

	h.schedulerUC.StartScheduler()
	return nil
}

func (h *SchedulerHandler) HandleInsertConnector(ctx context.Context, connectorID uint64) error {
	connector, settings, err := h.getConnectorConfig(ctx, connectorID)
	if err != nil {
		return err
	}

	return h.start(connector, settings)
}

func (h *SchedulerHandler) HandleUpdateConnector(ctx context.Context, connectorID uint64) error {
	connector, settings, err := h.getConnectorConfig(ctx, connectorID)
	if err != nil {
		return err
	}

	if !connector.Enabled || !settings.IncSync {
		err := h.schedulerUC.RemoveCronJob(connectorID)
		if err != nil {
			return err
		}
		return nil
	}

	currentPeriod, exists := h.schedulerUC.GetCronJobPeriod(connectorID)
	if exists {
		if currentPeriod == settings.IncSyncPeriod {
			return nil
		}
		err := h.schedulerUC.RemoveCronJob(connectorID)
		if err != nil {
			return err
		}
	}

	return h.start(connector, settings)
}

func (h *SchedulerHandler) getConnectorConfig(ctx context.Context, connectorID uint64) (*domain.Connector, *domain.SyncSettings, error) {
	connector, err := h.connectorUC.GetByID(ctx, connectorID)
	if err != nil {
		return nil, nil, err
	}

	settings, err := connector.GetSyncSettings()
	if err != nil {
		return nil, nil, err
	}
	return connector, settings, nil
}

func (h *SchedulerHandler) start(connector *domain.Connector, settings *domain.SyncSettings) error {
	message := &domain.QueueMessage{
		ConnectorID: connector.ConnectorID,
		TaskType:    domain.TaskTypeIncrementalSync,
		Queue:       domain.LowQueue,
	}

	err := h.schedulerUC.CreateCronJob(settings.IncSyncPeriod, message)
	if err != nil {
		return err
	}

	h.schedulerUC.StartScheduler()
	return nil
}
