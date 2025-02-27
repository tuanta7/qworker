package domain

type Message struct {
	ConnectorID uint64  `json:"connector_id"`
	JobType     JobType `json:"job_type"`
}
