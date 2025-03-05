package scheduleruc

import (
	redisrepo "github.com/tuanta7/qworker/internal/repository/redis"

	"github.com/tuanta7/qworker/internal/domain"
	pgrepo "github.com/tuanta7/qworker/internal/repository/postgres"
	"github.com/tuanta7/qworker/pkg/logger"
	"go.uber.org/zap"
)

type UseCase struct {
	schedulerRepo *pgrepo.ConnectorRepository
	jobRepo       *redisrepo.JobRepository
	logger        *logger.ZapLogger
}

func NewUseCase(
	schedulerRepo *pgrepo.ConnectorRepository,
	jobRepo *redisrepo.JobRepository,
	logger *logger.ZapLogger,
) *UseCase {
	return &UseCase{
		schedulerRepo: schedulerRepo,
		jobRepo:       jobRepo,
		logger:        logger,
	}
}

func (u *UseCase) GetEnabledConnectors() {}

func (u *UseCase) SendSyncJob(connectorID uint64) func() {
	return func() {
		message := domain.Message{
			ConnectorID: connectorID,
			JobType:     domain.JobTypeIncrementalSync,
		}

		payload, err := json.Marshal(message)
		if err != nil {
			return nil, err
		}

		task, err := u.jobRepo.Enqueue(asynq.NewTask(payload, domain.IncrementalSyncJobQueue))
		if err != nil {
			u.logger.Error(
				"SchedulerUsecase -  SendSyncMessage - u.jobRepo.Enqueue",
				zap.Error(err),
				zap.Uint64("connector_id", connectorID),
			)
		}

		u.logger.Info(
			"SchedulerUsecase - SendSyncMessage - OK",
			zap.Any("task", task),
		)
	}
}
