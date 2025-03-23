package domain

import (
	"context"
	"time"
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
