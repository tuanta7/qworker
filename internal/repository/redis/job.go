package redisrepo

import (
	"encoding/json"
	"github.com/hibiken/asynq"
)

type JobRepository struct {
	client *asynq.Client
}

func NewJobRepository(client *asynq.Client) *JobRepository {
	return &JobRepository{client}
}

func (r *JobRepository) Enqueue(task *asynq.Task) (*asynq.TaskInfo, error) {
	return r.client.Enqueue(task)
}
