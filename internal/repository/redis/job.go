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

func (r *JobRepository) Enqueue(queueName string, message any) (*asynq.TaskInfo, error) {
	payload, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	info, err := r.client.Enqueue(asynq.NewTask(queueName, payload))
	if err != nil {
		return nil, err
	}

	return info, nil
}
