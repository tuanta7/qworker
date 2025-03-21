package domain

import (
	"context"
	"time"
)

const (
	TaskTypeIncrementalSync = "user:incremental_sync"
	TaskTypeFullSync        = "user:full_sync"
	TaskTypeTerminate       = "user:sync_terminate"
)

type QueueMessage struct {
	ConnectorID uint64 `json:"connector_id"`
	TaskType    string `json:"task_type"`
	Queue       string `json:"queue"`
}

type Task struct {
	Type      string
	StartedAt time.Time
	Cancel    context.CancelFunc
}
