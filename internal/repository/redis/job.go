package redisrepo

import (
	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/pkg/db"
	"time"
)

type JobRepository struct {
	asynqClient *asynq.Client
	*db.RedisClient
}

func NewJobRepository(asynqClient *asynq.Client) *JobRepository {
	return &JobRepository{
		asynqClient: asynqClient,
	}
}

func (r *JobRepository) Enqueue(task *asynq.Task) (*asynq.TaskInfo, error) {
	opts := []asynq.Option{
		asynq.MaxRetry(5),
		asynq.Retention(5 * time.Minute),
	}

	return r.asynqClient.Enqueue(task, opts...)
}
