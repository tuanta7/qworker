package domain

import (
	"database/sql/driver"
	"encoding/json"
	"reflect"
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
		SS SyncSettings `json:"syncSettings"`
	}{}

	err := json.Unmarshal(c.Data.Raw, &data)
	if err != nil {
		return nil, err
	}
	return &data.SS, nil
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

type SyncSettings struct {
	BatchSize     uint32        `json:"batchSize"`
	IncSync       bool          `json:"incrementalSyncEnabled"`
	IncSyncPeriod time.Duration `json:"incrementalSyncPeriod"`
}

type Mapper struct {
	ExternalID  string            `json:"external_id"`
	Username    string            `json:"username"`
	FullName    string            `json:"full_name"`
	Email       string            `json:"email"`
	PhoneNumber string            `json:"phone_number"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
	Custom      map[string]string `json:"custom"`
}

func (m *Mapper) Scan(v any) error {
	if v == nil {
		return nil
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		return nil
	}

	var data []byte
	switch val := v.(type) {
	case string:
		data = []byte(val)
	case []byte:
		data = val
	default:
		return nil
	}

	if !json.Valid(data) {
		return nil
	}

	return json.Unmarshal(data, m)
}

func (m *Mapper) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *Mapper) ToMap() (map[string]string, error) {
	return nil, nil
}
