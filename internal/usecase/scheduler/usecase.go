package scheduleruc

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/robfig/cron/v3"
	"github.com/tuanta7/qworker/config"
	"github.com/tuanta7/qworker/internal/domain"
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
	jobs           map[string]JobInfo // expect to have only one scheduler instance
	cronScheduler  *cron.Cron
	asynqClient    *asynq.Client
	asynqInspector *asynq.Inspector
	logger         *logger.ZapLogger
}

func NewUseCase(
	asynqClient *asynq.Client,
	asynqInspector *asynq.Inspector,
	logger *logger.ZapLogger,
) *UseCase {
	return &UseCase{
		lock:           sync.Mutex{},
		cronScheduler:  cron.New(cron.WithSeconds()),
		jobs:           make(map[string]JobInfo),
		asynqClient:    asynqClient,
		asynqInspector: asynqInspector,

		logger: logger,
	}
}

func (u *UseCase) GetJobPeriod(connectorID string) (time.Duration, bool) {
	u.lock.Lock()
	jobInfo, exists := u.jobs[connectorID]
	u.lock.Unlock()

	if !exists {
		return 0, false
	}

	return jobInfo.Period, true
}

func (u *UseCase) CleanJob(connectorID string) error {
	u.lock.Lock()
	jobInfo, exists := u.jobs[connectorID]
	u.lock.Unlock()

	if exists {
		u.cronScheduler.Remove(jobInfo.EntryID)
		delete(u.jobs, connectorID)
		u.logger.Info("Job removed", zap.Any("id", jobInfo.EntryID))
	}

	var errs []error
	for q := range config.QueuePriority {
		err := u.asynqInspector.DeleteTask(q, connectorID)
		if err != nil {
			if errors.Is(err, asynq.ErrQueueNotFound) || errors.Is(err, asynq.ErrTaskNotFound) {
				continue
			}
			u.logger.Warn("u.asynqInspector.DeleteTask", zap.Error(err))
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func (u *UseCase) CreateJob(period time.Duration, queue string, message *domain.QueueMessage) error {
	spec := fmt.Sprintf("@every %s", (period * time.Second).String())
	cmd := u.enqueueTaskCMD(message, queue)

	jobID, err := u.cronScheduler.AddFunc(spec, cmd)
	if err != nil {
		return err
	}

	u.lock.Lock()
	u.jobs[strconv.FormatUint(message.ConnectorID, 10)] = JobInfo{
		EntryID: jobID,
		Period:  period,
	}
	defer u.lock.Unlock()

	u.logger.Info("Create cron job", zap.Any("message", message))
	return nil
}

func (u *UseCase) enqueueTaskCMD(message *domain.QueueMessage, queue string) func() {
	return func() {
		taskID := strconv.FormatUint(message.ConnectorID, 10)
		payload, err := json.Marshal(message)
		if err != nil {
			u.logger.Error("SchedulerUsecase -  enqueueTaskCMD - json.Marshal", zap.Error(err))
			return
		}

		ok, err := u.IsTaskAllowed(taskID)
		if err != nil || !ok {
			u.logger.Error("SchedulerUsecase -  enqueueTaskCMD - u.IsTaskAllowed",
				zap.String("message", "this task is not allowed to be enqueued right now"),
				zap.String("type", message.TaskType),
				zap.Bool("allowed", ok),
				zap.Error(err))
			return
		}

		taskInfo, _ := u.asynqInspector.GetTaskInfo(queue, taskID)
		if taskInfo != nil && taskInfo.State == asynq.TaskStateArchived {
			_ = u.asynqInspector.DeleteTask(queue, taskID)
		}

		task, err := u.asynqClient.Enqueue(
			asynq.NewTask(message.TaskType, payload),
			asynq.TaskID(taskID),
			asynq.Queue(queue),
			asynq.MaxRetry(0),
			asynq.Retention(0),
		)
		if err != nil {
			u.logger.Error("SchedulerUsecase -  enqueueTaskCMD - u.asynqClient.Enqueue", zap.Error(err))
			return
		}

		u.logger.Info("enqueue new task", zap.Any("task", task))
	}
}

// IsTaskAllowed checks whether a task can be enqueued. It does not guarantee precise elimination of
// conflicting tasks but ensures conflict prevention on scheduler side in a strict and extreme manner.
func (u *UseCase) IsTaskAllowed(id string) (bool, error) {
	fullSyncTask, err := u.asynqInspector.GetTaskInfo(config.QueueFullSync, id)
	if err != nil {
		if errors.Is(err, asynq.ErrTaskNotFound) || errors.Is(err, asynq.ErrQueueNotFound) {
			return true, nil
		}
		return false, err
	}

	if fullSyncTask != nil {
		return fullSyncTask.State == asynq.TaskStateArchived, nil
	}

	return true, nil
}

func (u *UseCase) ClearAllJobs() {
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
