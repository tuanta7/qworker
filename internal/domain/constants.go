package domain

type JobType string

const (
	SyncJobQueueName = "user:sync"

	JobTypeIncrementalSync JobType = "INCREMENTAL_SYNC"
	JobTypeFullSync        JobType = "FULL_SYNC"
)
