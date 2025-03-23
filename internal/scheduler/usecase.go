package scheduleruc

import (
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/robfig/cron/v3"
	"github.com/tuanta7/qworker/internal/domain"
	redisrepo "github.com/tuanta7/qworker/internal/repository/redis"
	"github.com/tuanta7/qworker/pkg/logger"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"time"
)

type JobInfo struct {
	EntryID cron.EntryID
	Period  time.Duration
}

type UseCase struct {
	lock           sync.Mutex
	cronScheduler  *cron.Cron
	jobs           map[uint64]*JobInfo
	asynqClient    *asynq.Client
	asynqInspector *asynq.Inspector
	taskRepository *redisrepo.TaskRepository
	logger         *logger.ZapLogger
}

func NewUseCase(
	asynqClient *asynq.Client,
	asynqInspector *asynq.Inspector,
	taskRepository *redisrepo.TaskRepository,
	logger *logger.ZapLogger,
) *UseCase {
	return &UseCase{
		lock:           sync.Mutex{},
		cronScheduler:  cron.New(cron.WithSeconds()),
		jobs:           make(map[uint64]*JobInfo), // expect to have only one scheduler instance
		asynqClient:    asynqClient,
		asynqInspector: asynqInspector,
		taskRepository: taskRepository,
		logger:         logger,
	}
}

func (u *UseCase) GetCronJobPeriod(connectorID uint64) (time.Duration, bool) {
	u.lock.Lock()
	jobInfo, exists := u.jobs[connectorID]
	if !exists {
		return 0, false
	}
	u.lock.Unlock()

	return jobInfo.Period, true
}

func (u *UseCase) CreateCronJob(period time.Duration, message *domain.QueueMessage) error {
	spec := fmt.Sprintf("@every %s", (period * time.Second).String())
	cmd := u.tryEnqueueTask(message, message.Queue)

	jobID, err := u.cronScheduler.AddFunc(spec, cmd)
	if err != nil {
		return err
	}

	u.lock.Lock()
	u.jobs[message.ConnectorID] = &JobInfo{
		EntryID: jobID,
		Period:  period,
	}
	u.lock.Unlock()

	return nil
}

func (u *UseCase) tryEnqueueTask(message *domain.QueueMessage, queue string) func() {
	return func() {
		payload, _ := json.Marshal(message)

		task, err := u.asynqClient.Enqueue(
			asynq.NewTask(message.TaskType, payload),
			asynq.TaskID(strconv.FormatUint(message.ConnectorID, 10)),
			asynq.Queue(queue),
			asynq.MaxRetry(0),
			asynq.Retention(0),
		)
		if err != nil {
			u.logger.Error(
				"SchedulerUsecase -  tryEnqueueTask - u.asynqClient.Enqueue",
				zap.Error(err),
				zap.Any("payload", payload),
			)
		}

		u.logger.Info("enqueue new task", zap.Any("task", task))
	}
}

func (u *UseCase) RemoveCronJob(connectorID uint64) error {
	u.lock.Lock()
	jobInfo, exists := u.jobs[connectorID]
	if exists {
		u.cronScheduler.Remove(jobInfo.EntryID)
	}
	u.lock.Unlock()

	// TODO: Stop processing/pending tasks
	err := u.asynqInspector.CancelProcessing(strconv.FormatUint(connectorID, 10))
	if err != nil {
		return err
	}

	return nil
}

func (u *UseCase) ClearCronJobs() {
	u.cronScheduler.Stop()
	for _, e := range u.cronScheduler.Entries() {
		u.cronScheduler.Remove(e.ID)
	}

	u.lock.Lock()
	clear(u.jobs)
	u.lock.Unlock()
}

func (u *UseCase) StartScheduler() {
	u.cronScheduler.Start()
}
