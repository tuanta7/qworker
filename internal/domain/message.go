package domain

import (
	"context"
	"time"
)

type JobType string

const (
	JobTypeIncrementalSync JobType = "INCREMENTAL_SYNC"
	JobTypeFullSync        JobType = "FULL_SYNC"
	JobTypeTerminate       JobType = "TERMINATE"
)

type Message struct {
	ConnectorID uint64  `json:"connector_id"`
	JobType     JobType `json:"job_type"`
}

type Job struct {
	JobType   JobType
	StartedAt time.Time
	Cancel    context.CancelFunc
}
