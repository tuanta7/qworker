package domain

type JobType string

const (
	IncrementalSyncJobQueue = "user:incremental_sync"
	FullSyncJobQueue        = "user:full_sync"

	JobTypeIncrementalSync JobType = "INCREMENTAL_SYNC"
	JobTypeFullSync        JobType = "FULL_SYNC"

	TableConnectors  = "connectors"
	ColConnectorID   = "connector_id"
	ColConnectorType = "connector_type"
	ColDisplayName   = "display_name"
	ColEnabled       = "enabled"
	ColLastSync      = "last_sync"

	ColData      = "data"
	ColCreatedAt = "created_at"
	ColUpdatedAt = "updated_at"
)

var (
	AllConnectorCols = []string{
		ColConnectorID,
		ColConnectorType,
		ColDisplayName,
		ColEnabled,
		ColLastSync,
		ColData,
		ColCreatedAt,
		ColUpdatedAt,
	}
)
