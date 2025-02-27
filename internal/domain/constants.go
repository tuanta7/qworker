package domain

type JobType string

const (
	JobTypeIncrementalSync JobType = "INCREMENTAL_SYNC"
	JobTypeFullSync        JobType = "FULL_SYNC"
)
