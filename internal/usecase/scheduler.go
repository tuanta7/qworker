package usecase

import (
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/internal/domain"
	pgrepo "github.com/tuanta7/qworker/internal/repository/postgres"
	"github.com/tuanta7/qworker/pkg/logger"
	"go.uber.org/zap"
)

type SchedulerUsecase struct {
	schedulerRepo *pgrepo.SchedulerRepository
	logger        *logger.ZapLogger
	asynqClient   *asynq.Client
}

func NewSchedulerUsecase(schedulerRepo *pgrepo.SchedulerRepository, logger *logger.ZapLogger, asynqClient *asynq.Client) *SchedulerUsecase {
	return &SchedulerUsecase{
		schedulerRepo: schedulerRepo,
		logger:        logger,
		asynqClient:   asynqClient,
	}
}

func (u *SchedulerUsecase) SendSyncMessage(connectorID uint64) func() {
	return func() {
		u.logger.Info("SchedulerUsecase - SendSyncMessage")

		message := domain.Message{
			ConnectorID: connectorID,
			JobType:     domain.JobTypeIncrementalSync,
		}

		payload, err := json.Marshal(message)
		if err != nil {
			u.logger.Error(
				"SchedulerUsecase -  SendSyncMessage - json.Marshal",
				zap.Error(err),
				zap.Uint64("connector_id", connectorID),
			)
		}

		info, err := u.asynqClient.Enqueue(asynq.NewTask("user:sync", payload))
		if err != nil {
			u.logger.Error(
				"SchedulerUsecase - SendSyncMessage - asynqClient.Enqueue",
				zap.Error(err),
				zap.Uint64("connector_id", connectorID),
			)
		}

		u.logger.Info(
			"SchedulerUsecase - SendSyncMessage",
			zap.Any("info", info),
		)
	}
}
