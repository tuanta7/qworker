package scheduleruc

import (
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/robfig/cron/v3"
	"github.com/tuanta7/qworker/config"
	"github.com/tuanta7/qworker/internal/domain"
	"github.com/tuanta7/qworker/pkg/logger"
	"go.uber.org/zap"
	"sync"
	"time"
)

type JobInfo struct {
	EntryID cron.EntryID
	Period  time.Duration
}

type UseCase struct {
	lock        sync.Mutex
	scheduler   *cron.Cron
	jobs        map[uint64]*JobInfo
	asynqClient *asynq.Client
	logger      *logger.ZapLogger
}

func NewUseCase(asynqClient *asynq.Client, logger *logger.ZapLogger) *UseCase {
	return &UseCase{
		lock:        sync.Mutex{},
		scheduler:   cron.New(cron.WithSeconds()),
		jobs:        make(map[uint64]*JobInfo),
		asynqClient: asynqClient,
		logger:      logger,
	}
}

func (u *UseCase) GetJobPeriod(connectorID uint64) (time.Duration, bool) {
	u.lock.Lock()
	defer u.lock.Unlock()

	jobInfo, exists := u.jobs[connectorID]
	if !exists {
		return 0, false
	}

	return jobInfo.Period, true
}

func (u *UseCase) CreateJob(message *domain.QueueMessage, period time.Duration, queue string) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	u.lock.Lock()
	defer u.lock.Unlock()

	typename := queue
	cmd := u.enqueueTask(asynq.NewTask(typename, payload), queue)
	jobID, err := u.scheduler.AddFunc(fmt.Sprintf("@every %s", (period*time.Second).String()), cmd)
	if err != nil {
		return err
	}

	u.jobs[message.ConnectorID] = &JobInfo{
		EntryID: jobID,
		Period:  period,
	}
	return nil
}

func (u *UseCase) RemoveJob(connectorID uint64) {
	u.lock.Lock()

	jobInfo, exists := u.jobs[connectorID]
	if exists {
		u.scheduler.Remove(jobInfo.EntryID)
	}

	u.lock.Unlock()

	payload, _ := json.Marshal(&domain.QueueMessage{
		ConnectorID: connectorID,
		TaskType:    domain.TaskTypeTerminate,
	})
	u.enqueueTask(asynq.NewTask("terminate", payload), config.TerminateQueue)()
}

func (u *UseCase) StartScheduler() {
	u.scheduler.Start()
}

func (u *UseCase) RemoveJobs() {
	u.scheduler.Stop()
	for _, e := range u.scheduler.Entries() {
		u.scheduler.Remove(e.ID)
	}

	u.lock.Lock()
	clear(u.jobs)
	u.lock.Unlock()
}

func (u *UseCase) enqueueTask(task *asynq.Task, queue string) func() {
	return func() {
		t, err := u.asynqClient.Enqueue(task,
			asynq.MaxRetry(0),
			asynq.Retention(5*time.Minute),
			asynq.Queue(queue))
		if err != nil {
			u.logger.Error(
				"SchedulerUsecase -  enqueueTask - u.asynqClient.Enqueue",
				zap.Error(err),
				zap.Any("task", task.Payload),
			)
		}

		u.logger.Info("enqueue new task", zap.Any("task", t))
	}
}
