package domain

type CronJob struct {
	JobID       uint64 `json:"job_id"`
	ConnectorID uint64 `json:"connector_id"`
	JobType     string `json:"job_type"`
	Status      string `json:"status"`
}
