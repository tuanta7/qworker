package connectoruc

import (
	"context"
	"github.com/tuanta7/qworker/internal/domain"
)

type Repository interface {
	ListByEnabled(ctx context.Context, enabled bool) ([]*domain.Connector, error)
	GetByID(ctx context.Context, id uint64) (*domain.Connector, error)
}
