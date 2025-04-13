package handler

import (
	"context"
	"github.com/tuanta7/qworker/config"
	"github.com/tuanta7/qworker/internal/domain"
	"github.com/tuanta7/qworker/internal/usecase/connector"
	"github.com/tuanta7/qworker/internal/usecase/scheduler"
	"strconv"
	"time"
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

func (h *SchedulerHandler) Init(ctx context.Context) error {
	connectors, err := h.connectorUC.ListEnabled(ctx)
	if err != nil {
		return err
	}

	for _, connector := range connectors {
		settings, err := connector.GetSyncSettings()
		if err != nil || !settings.IncSync {
			continue
		}

		err = h.createIncrementalSyncJob(connector.ConnectorID, settings.IncSyncPeriod)
		if err != nil {
			return err
		}
	}

	h.schedulerUC.StartScheduler()
	return nil
}

func (h *SchedulerHandler) Clear() {
	h.schedulerUC.ClearAllJobs()
}

func (h *SchedulerHandler) HandleInsertConnector(ctx context.Context, message *domain.NotifyMessage) error {
	connector, err := h.connectorUC.GetByID(ctx, message.ID)
	if err != nil {
		return err
	}

	if !connector.Enabled {
		return nil
	}

	syncSettings, err := connector.GetSyncSettings()
	if err != nil {
		return err
	}

	if !syncSettings.IncSync {
		return nil
	}

	err = h.createIncrementalSyncJob(connector.ConnectorID, syncSettings.IncSyncPeriod)
	if err != nil {
		return err
	}
	return nil
}

func (h *SchedulerHandler) HandleUpdateConnector(ctx context.Context, message *domain.NotifyMessage) error {
	connector, err := h.connectorUC.GetByID(ctx, message.ID)
	if err != nil {
		return err
	}

	syncSettings, err := connector.GetSyncSettings()
	if err != nil {
		return err
	}

	sID := strconv.FormatUint(message.ID, 10)
	if !connector.Enabled || !syncSettings.IncSync {
		return h.schedulerUC.CleanJob(sID)
	}

	currentPeriod, exists := h.schedulerUC.GetJobPeriod(sID)
	if exists {
		if currentPeriod == syncSettings.IncSyncPeriod {
			return nil
		}
		_ = h.schedulerUC.CleanJob(sID)
	}

	err = h.createIncrementalSyncJob(connector.ConnectorID, syncSettings.IncSyncPeriod)
	if err != nil {
		return err
	}
	return nil
}

func (h *SchedulerHandler) HandleDeleteConnector(ctx context.Context, message *domain.NotifyMessage) error {
	return h.schedulerUC.CleanJob(strconv.FormatUint(message.ID, 10))
}

func (h *SchedulerHandler) createIncrementalSyncJob(id uint64, period time.Duration) error {
	queue := config.QueueIncrementalSync
	taskType := config.QueueTask[queue]

	err := h.schedulerUC.CreateJob(period, queue, &domain.QueueMessage{
		ConnectorID: id,
		TaskType:    taskType,
	})
	if err != nil {
		return err
	}

	return nil
}
