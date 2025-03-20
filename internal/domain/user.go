package domain

import (
	"github.com/tuanta7/qworker/pkg/sqlxx"
	"time"
)

type User struct {
	UserID        string         `json:"id"`
	Username      string         `json:"username"`
	FullName      string         `json:"fullName"`
	PhoneNumber   string         `json:"phoneNumber"`
	Email         string         `json:"email"`
	EmailVerified bool           `json:"emailVerified"`
	Active        bool           `json:"active"`
	SourceID      *uint64        `json:"sourceID"`
	Data          sqlxx.TextData `json:"data"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

const (
	TableUser        string = "private.user"
	ColUserID        string = "id"
	ColUsername      string = "username"
	ColFullName      string = "full_name"
	ColPhoneNumber   string = "phone_number"
	ColEmail         string = "email"
	ColEmailVerified string = "email_verified"
	ColActive        string = "active"
	ColSourceID      string = "source_id"
)

var (
	AllUserCols = []string{
		ColUserID,
		ColUsername,
		ColFullName,
		ColPhoneNumber,
		ColEmail,
		ColEmailVerified,
		ColActive,
		ColSourceID,
		ColData,
		ColCreatedAt,
		ColUpdatedAt,
	}

	AllUserSyncCols = []string{
		ColUserID,
		ColUsername,
		ColFullName,
		ColPhoneNumber,
		ColEmail,
		ColSourceID,
		ColData,
		ColCreatedAt,
		ColUpdatedAt,
	}
)
