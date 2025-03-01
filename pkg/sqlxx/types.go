package sqlxx

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type TextData struct {
	Raw    []byte
	Parsed any
}

// implement the sql.Scanner interface
func (t *TextData) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		t.Raw = v
		return nil
	case string:
		t.Raw = []byte(v)
		return nil
	}

	return errors.New("error while scanning TextData: unsupported type")
}

// implement the driver.Valuer interface
func (t TextData) Value() (driver.Value, error) {
	return json.Marshal(t.Parsed)
}

// implement the json.Marshaler interface
func (t TextData) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Parsed)
}
