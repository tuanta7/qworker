package domain

import "fmt"

const (
	CriticalQueue = "critical"
	DefaultQueue  = "default"
	LowQueue      = "low"

	TaskTypeIncrementalSync = "user:incremental_sync"
	TaskTypeFullSync        = "user:full_sync"
	TaskTypeTerminate       = "user:sync_terminate"
)

var (
	QueuePriority = map[string]int{
		CriticalQueue: 6,
		DefaultQueue:  3,
		LowQueue:      1,
	}

	TaskPriority = map[string]int{
		TaskTypeTerminate:       3,
		TaskTypeFullSync:        2,
		TaskTypeIncrementalSync: 1,
	}
)

const (
	ColData      = "data"
	ColCreatedAt = "created_at"
	ColUpdatedAt = "updated_at"
)

func FormatTaskID(taskType string, connectorID uint64) string {
	return fmt.Sprintf("%s-%d", taskType, connectorID)
}
