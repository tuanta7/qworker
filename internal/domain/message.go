package domain

type JobType string

type Message struct {
	ConnectorID uint64  `json:"connector_id"`
	JobType     JobType `json:"job_type"`
}

const (
	JobTypeIncrementalSync JobType = "INCREMENTAL_SYNC"
	JobTypeFullSync        JobType = "FULL_SYNC"
)
