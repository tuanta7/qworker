package jobmanager

import (
	"errors"
	"github.com/tuanta7/qworker/internal/domain"
	"sync"
	"time"
)

type Job struct {
	Type      domain.JobType
	Running   bool
	StartedAt time.Time
}

type JobManager struct {
	lock  *sync.RWMutex
	tasks map[uint64]*Job
}

func NewJobManager() *JobManager {
	return &JobManager{
		lock:  new(sync.RWMutex),
		tasks: make(map[uint64]*Job),
	}
}

func (m *JobManager) GetJobStatus(connectorID uint64) (bool, error) {
	m.lock.RLock()
	task, exist := m.tasks[connectorID]
	m.lock.RUnlock()

	if exist {
		return task.Running, nil
	}

	return false, errors.New("task not found")
}

func (m *JobManager) CreateJob() (*Job, error) {
	return nil, nil
}

func (m *JobManager) RunJob(id string) error {
	return nil
}

func (m *JobManager) TerminateJob() error {
	return nil
}
