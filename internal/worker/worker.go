package workeruc

import (
	pgrepo "github.com/tuanta7/qworker/internal/repository/postgres"
	"github.com/tuanta7/qworker/pkg/logger"
)

type UseCase struct {
	connectorRepository *pgrepo.ConnectorRepository
	logger              *logger.ZapLogger
}

func NewUseCase(connectorRepository *pgrepo.ConnectorRepository, logger *logger.ZapLogger) *UseCase {
	return &UseCase{
		connectorRepository: connectorRepository,
		logger:              logger,
	}
}
