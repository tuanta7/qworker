package domain

import (
	"context"
	"time"
)

type TaskType string

const (
	TaskTypeIncrementalSync TaskType = "INCREMENTAL_SYNC"
	TaskTypeFullSync        TaskType = "FULL_SYNC"
	TaskTypeTerminate       TaskType = "TERMINATE"
)

type QueueMessage struct {
	ConnectorID uint64   `json:"connector_id"`
	TaskType    TaskType `json:"task_type"`
}

type Task struct {
	Type      TaskType
	StartedAt time.Time
	Cancel    context.CancelFunc
}
