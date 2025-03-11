package workeruc

import (
	"context"
	"github.com/tuanta7/qworker/internal/domain"
	pgrepo "github.com/tuanta7/qworker/internal/repository/postgres"
	"github.com/tuanta7/qworker/pkg/logger"
	"github.com/tuanta7/qworker/pkg/utils"
	"sync"
	"time"
)

type UseCase struct {
	lock                *sync.Mutex
	runningJobs         map[uint64]*domain.Job
	connectorRepository *pgrepo.ConnectorRepository
	logger              *logger.ZapLogger
}

func NewUseCase(connectorRepository *pgrepo.ConnectorRepository, zl *logger.ZapLogger) *UseCase {
	return &UseCase{
		lock:                new(sync.Mutex),
		runningJobs:         make(map[uint64]*domain.Job),
		connectorRepository: connectorRepository,
		logger:              zl,
	}
}

func (u *UseCase) GetJob(connectorID uint64) (*domain.Job, error) {
	u.lock.Lock()
	job, exist := u.runningJobs[connectorID]
	u.lock.Unlock()

	if exist {
		return job, nil
	}

	return nil, utils.ErrJobNotFound
}

func (u *UseCase) RunJob(ctx context.Context, message domain.Message) error {
	c, cancel := context.WithCancel(ctx)

	u.lock.Lock()
	u.runningJobs[message.ConnectorID] = &domain.Job{
		JobType:   message.JobType,
		Cancel:    cancel,
		StartedAt: time.Now(),
	}
	u.lock.Unlock()

	defer delete(u.runningJobs, message.ConnectorID)

	done := make(chan error)
	go func(m domain.Message) {
		defer close(done)

		err := u.sync(m)
		if err != nil {
			done <- err
		}
	}(message)

	select {
	case <-c.Done():
		return c.Err()
	case err := <-done: // also unblocked when done is closed
		return err
	}
}

func (u *UseCase) CancelJob(connectorID uint64) {
	u.lock.Lock()
	defer u.lock.Unlock()

	job, exists := u.runningJobs[connectorID]
	if exists && job.Cancel != nil {
		job.Cancel()
		delete(u.runningJobs, connectorID)
	}
}

func (u *UseCase) sync(message domain.Message) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	currJob, exists := u.runningJobs[message.ConnectorID]

	switch message.JobType {
	case domain.JobTypeIncrementalSync:
		if exists {
			// Ignore
			return nil
		}
		// Run Incremental Sync
	case domain.JobTypeFullSync:
		if exists && currJob.JobType == domain.JobTypeIncrementalSync {
			// Run Full Sync
			return nil
		}
		return nil
	case domain.JobTypeTerminate:
		u.CancelJob(message.ConnectorID)
	}

	return nil
}
