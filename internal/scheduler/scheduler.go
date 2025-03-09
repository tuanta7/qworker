package scheduleruc

import (
	"github.com/hibiken/asynq"
	redisrepo "github.com/tuanta7/qworker/internal/repository/redis"

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

func (u *UseCase) EnqueueTask(task *asynq.Task) func() {
	return func() {
		task, err := u.jobRepo.Enqueue(task)
		if err != nil {
			u.logger.Error(
				"SchedulerUsecase -  SendSyncMessage - u.jobRepo.Enqueue",
				zap.Error(err),
				zap.Any("task", task.Payload),
			)
		}

		u.logger.Info(
			"SchedulerUsecase - SendSyncMessage - OK",
			zap.Any("task", task),
		)
	}
}
