package domain

import (
	"encoding/json"
	"github.com/tuanta7/qworker/pkg/sqlxx"
	"time"
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
		Settings SyncSettings `json:"syncSettings"`
	}{}

	err := json.Unmarshal(c.Data.Raw, &data)
	if err != nil {
		return nil, err
	}

	return &data.Settings, nil
}

type LDAPConnector struct {
	URL                   string        `json:"url"`
	ConnectTimeout        time.Duration `json:"connectTimeout"`
	ReadTimeout           time.Duration `json:"readTimeout"`
	SystemAccountDN       string        `json:"systemAccountDn"`
	SystemAccountPassword string        `json:"systemAccountPassword"`
	UsernameAttribute     string        `json:"usernameAttribute"`
	BaseDN                string        `json:"baseDn"`
	SyncSettings          SyncSettings  `json:"syncSettings"`
}

type SCIMConnector struct {
	BaseURL      string       `json:"baseUrl"`
	SyncSettings SyncSettings `json:"syncSettings"`
}

type SyncSettings struct {
	BatchSize     uint32        `json:"batchSize"`
	IncSync       bool          `json:"incrementalSyncEnabled"`
	IncSyncPeriod time.Duration `json:"incrementalSyncPeriod"`
}
