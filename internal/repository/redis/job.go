package redisrepo

import (
	"github.com/hibiken/asynq"
	"time"
)

type JobRepository struct {
	asynqClient *asynq.Client
}

func NewJobRepository(asynqClient *asynq.Client) *JobRepository {
	return &JobRepository{
		asynqClient: asynqClient,
	}
}

func (r *JobRepository) Enqueue(task *asynq.Task) (*asynq.TaskInfo, error) {
	opts := []asynq.Option{
		asynq.MaxRetry(1),
		asynq.Retention(5 * time.Minute), // stored as completed task for 5min
	}

	return r.asynqClient.Enqueue(task, opts...)
}
