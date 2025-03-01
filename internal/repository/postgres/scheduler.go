package pgrepo

import (
	"github.com/tuanta7/qworker/internal/domain"
	"github.com/tuanta7/qworker/pkg/db"
)

type SchedulerRepository struct {
	*db.PostgresClient
}

func NewSchedulerRepository(pc *db.PostgresClient) *SchedulerRepository {
	return &SchedulerRepository{pc}
}

func (r *SchedulerRepository) GetConnectors() []*domain.Connector {
	connectors := make([]*domain.Connector, 0)
	return connectors
}
