package pgrepo

import "github.com/tuanta7/qworker/pkg/db"

type SchedulerRepository struct {
	*db.PostgresClient
}

func NewSchedulerRepository(pc *db.PostgresClient) *SchedulerRepository {
	return &SchedulerRepository{pc}
}

func (r *SchedulerRepository) LoadConfiguration() {}
