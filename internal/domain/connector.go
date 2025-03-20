package domain

import (
	"encoding/json"
	"time"

	"github.com/tuanta7/qworker/pkg/sqlxx"
)

type ConnectorType string

const (
	ConnectorTypeLDAP ConnectorType = "ldap"
	ConnectorTypeSCIM ConnectorType = "scim"
)

type Connector struct {
	ConnectorID   uint64         `json:"id"`
	ConnectorType ConnectorType  `json:"connectorType"`
	DisplayName   string         `json:"displayName"`
	LastSync      time.Time      `json:"lastSync"`
	Enabled       bool           `json:"enabled"`
	Data          sqlxx.TextData `json:"data"`
	Mapper        Mapper         `json:"mapper,omitempty"`
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

	return &data.SyncSettings, nil
}

const (
	TableConnector   = "private.connector"
	ColConnectorID   = "id"
	ColConnectorType = "connector_type"
	ColDisplayName   = "display_name"
	ColEnabled       = "enabled"
	ColLastSync      = "last_sync"
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
