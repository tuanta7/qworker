package domain

import (
	"database/sql/driver"
	"encoding/json"
	"reflect"
)

type Mapper struct {
	ExternalID  string            `json:"external_id"`
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
