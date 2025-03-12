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
	runningTask         map[uint64]*domain.Task
	connectorRepository *pgrepo.ConnectorRepository
	logger              *logger.ZapLogger
}

func NewUseCase(connectorRepository *pgrepo.ConnectorRepository, zl *logger.ZapLogger) *UseCase {
	return &UseCase{
		lock:                new(sync.Mutex),
		runningTask:         make(map[uint64]*domain.Task),
		connectorRepository: connectorRepository,
		logger:              zl,
	}
}

func (u *UseCase) GetJob(connectorID uint64) (*domain.Task, error) {
	u.lock.Lock()
	job, exist := u.runningTask[connectorID]
	u.lock.Unlock()

	if exist {
		return job, nil
	}

	return nil, utils.ErrJobNotFound
}

func (u *UseCase) RunJob(ctx context.Context, message domain.Message) error {
	c, cancel := context.WithCancel(ctx)

	u.lock.Lock()
	u.runningTask[message.ConnectorID] = &domain.Task{
		Type:      message.TaskType,
		Cancel:    cancel,
		StartedAt: time.Now(),
	}
	u.lock.Unlock()

	defer delete(u.runningTask, message.ConnectorID)

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

	job, exists := u.runningTask[connectorID]
	if exists && job.Cancel != nil {
		job.Cancel()
		delete(u.runningTask, connectorID)
	}
}

func (u *UseCase) sync(message domain.Message) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	currJob, exists := u.runningTask[message.ConnectorID]

	switch message.TaskType {
	case domain.TaskTypeIncrementalSync:
		if exists {
			// Ignore
			return nil
		}
		// Run Incremental Sync
	case domain.TaskTypeFullSync:
		if exists && currJob.Type == domain.TaskTypeIncrementalSync {
			// Run Full Sync
			return nil
		}
		return nil
	case domain.TaskTypeTerminate:
		u.CancelJob(message.ConnectorID)
	}

	return nil
}
