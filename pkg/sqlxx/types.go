package sqlxx

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type DurationString time.Duration

type TextData struct {
	Raw    []byte
	Parsed any
}

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

func (t TextData) Value() (driver.Value, error) {
	if t.Parsed == nil {
		return t.Raw, nil
	}

	return json.Marshal(t.Parsed)
}

func (t *TextData) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Parsed)
}
