package domain

import (
	"github.com/tuanta7/qworker/pkg/sqlxx"
)

type Mapping struct {
	ExternalID    string         `json:"external_id"`
	Email         string         `json:"email"`
	PhoneNumber   string         `json:"phone_number"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
	CustomMapping sqlxx.TextData `json:"data_text"`
}
