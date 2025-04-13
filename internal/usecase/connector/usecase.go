package connectoruc

import (
	"context"
	"github.com/tuanta7/qworker/internal/domain"
	"github.com/tuanta7/qworker/pkg/logger"
	"go.uber.org/zap"
)

type UseCase struct {
	connectorRepo Repository
	logger        *logger.ZapLogger
}

func NewUseCase(connectorRepo Repository, zl *logger.ZapLogger) *UseCase {
	return &UseCase{
		connectorRepo: connectorRepo,
		logger:        zl,
	}
}

func (u *UseCase) ListEnabled(ctx context.Context) ([]*domain.Connector, error) {
	connectors, err := u.connectorRepo.ListByEnabled(ctx, true)
	if err != nil {
		u.logger.Error(
			"Connector - UseCase - ListEnabledConnectors - u.connectorRepo.ListByEnabled",
			zap.Error(err))
		return nil, err
	}

	return connectors, nil
}

func (u *UseCase) GetByID(ctx context.Context, connectorID uint64) (*domain.Connector, error) {
	c, err := u.connectorRepo.GetByID(ctx, connectorID)
	if err != nil {
		u.logger.Error(
			"Connector - UseCase - GetByID - u.connectorRepo.GetByID",
			zap.Uint64("connector_id", connectorID),
			zap.Error(err))
		return nil, err
	}

	return c, nil
}
