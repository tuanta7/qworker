package config

const (
	TaskTypeIncrementalSync = "user:incremental_sync"
	TaskTypeFullSync        = "user:full_sync"

	QueueIncrementalSync = "inc"
	QueueFullSync        = "full"
)

var (
	QueuePriority = map[string]int{
		QueueFullSync:        3, // critical
		QueueIncrementalSync: 1, // default
	}

	QueueTask = map[string]string{
		QueueIncrementalSync: TaskTypeIncrementalSync,
		QueueFullSync:        TaskTypeFullSync,
	}
)
