package domain

import "time"

type Connector struct {
	ConnectorID   uint64    `json:"connector_id"`
	ConnectorName string    `json:"connector_name"`
	URL           string    `json:"url"`
	LastSync      time.Time `json:"last_sync"`
	SyncBatchSize int       `json:"sync_batch_size"`
	SyncPeriod    int       `json:"sync_period"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
