package domain

import (
	"encoding/json"
	"time"

	"github.com/tuanta7/qworker/pkg/sqlxx"
)

type ConnectorType string

const (
	ConnectorTypeLDAP ConnectorType = "LDAP"
)

type Connector struct {
	ConnectorID   uint64         `json:"id"`
	ConnectorType ConnectorType  `json:"connectorType"`
	DisplayName   string         `json:"displayName"`
	LastSync      time.Time      `json:"lastSync"`
	Enabled       bool           `json:"enabled"`
	Data          sqlxx.TextData `json:"data"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

func (c *Connector) GetSyncSettings() (*SyncSettings, error) {
	data := struct {
		SyncSettings SyncSettings `json:"syncSettings"`
	}{}

	err := json.Unmarshal(c.Data.Raw, &data)
	if err != nil {
		return nil, err
	}

	settings := &data.SyncSettings
	settings.IncSyncPeriod = settings.IncSyncPeriod * time.Second
	return settings, nil
}

type LdapConnector struct {
	URL            string        `json:"url"`
	ConnectTimeout time.Duration `json:"connectTimeout"`
	ReadTimeout    time.Duration `json:"readTimeout"`
	BindDN         string        `json:"bindDn"`
	BindPassword   string        `json:"bindPassword"`
	BaseDN         string        `json:"baseDn"`
	SearchScope    string        `json:"searchScope"`
	SyncSettings   SyncSettings  `json:"syncSettings"`
}

type SyncSettings struct {
	BatchSize     uint64        `json:"batchSize"`
	IncSync       bool          `json:"incrementalSyncEnabled"`
	IncSyncPeriod time.Duration `json:"incrementalSyncPeriod"`
}

const (
	TableConnector   = "connector"
	ColConnectorID   = "id"
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
