package domain

import (
	"time"

	"github.com/tuanta7/qworker/pkg/sqlxx"
)

type Connector struct {
	ConnectorID   uint64         `json:"connector_id"`
	ConnectorType string         `json:"connector_type"`
	ConnectorName string         `json:"connector_name"`
	LastSync      time.Time      `json:"last_sync"`
	Enabled       bool           `json:"enabled"`
	Data          sqlxx.TextData `json:"data"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type LdapConnector struct {
	URL            string        `json:"url"`
	ConnectTimeout time.Duration `json:"connect_timeout"`
	ReadTimeout    time.Duration `json:"read_timeout"`
	BindDN         string        `json:"bind_dn"`
	BindPassword   string        `json:"bind_password"`
	BaseDN         string        `json:"base_dn"`
	SearchScope    string        `json:"search_scope"`
	SyncSettings   SyncSettings  `json:"sync_settings"`
}

type SyncSettings struct {
	BatchSize   uint64        `json:"batch_size"`
	Incremental bool          `json:"incremental"`
	Period      time.Duration `json:"period"`
}
