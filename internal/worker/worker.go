package workeruc

import (
	"github.com/tuanta7/qworker/internal/domain"
	pgrepo "github.com/tuanta7/qworker/internal/repository/postgres"
	"github.com/tuanta7/qworker/pkg/jobmanager"
	"github.com/tuanta7/qworker/pkg/logger"
)

type UseCase struct {
	connectorRepository *pgrepo.ConnectorRepository
	jobManager          *jobmanager.JobManager
	logger              *logger.ZapLogger
}

func NewUseCase(connectorRepository *pgrepo.ConnectorRepository, logger *logger.ZapLogger) *UseCase {
	return &UseCase{
		connectorRepository: connectorRepository,
		logger:              logger,
	}
}

func (u *UseCase) CreateJob(connectorID uint64, jobType domain.JobType) error {
	return nil
}
