package domain

import (
	"context"
	"time"
)

type NotifyMessage struct {
	Table  string `json:"table"`
	Action string `json:"action"`
	ID     uint64 `json:"id"`
}

type QueueMessage struct {
	ConnectorID uint64 `json:"connector_id"`
	TaskType    string `json:"task_type"`
}

type Task struct {
	Type      string
	StartedAt time.Time
	Cancel    context.CancelFunc
}
