package usecase

import "github.com/hibiken/asynq"

func NewSyncJob() *asynq.Task {
	return asynq.NewTask("sync", nil)
}
