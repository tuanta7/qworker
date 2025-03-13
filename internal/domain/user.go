package domain

import "github.com/tuanta7/qworker/pkg/sqlxx"

type User struct {
	UserID        string         `json:"id"`
	DisplayName   string         `json:"displayName"`
	Email         string         `json:"email"`
	VerifiedEmail bool           `json:"verifiedEmail"`
	PhoneNumber   string         `json:"phoneNumber"`
	Data          sqlxx.TextData `json:"data"`
	CreatedAt     string         `json:"createdAt"`
	UpdatedAt     string         `json:"updatedAt"`
}

const (
	TableUser        string = "private.user"
	ColUserID        string = "id"
	ColEmail         string = "email"
	ColVerifiedEmail string = "verified_email"
	ColPhoneNumber   string = "phone_number"
)

var (
	AllUserCols = []string{
		ColUserID,
		ColDisplayName,
		ColEmail,
		ColVerifiedEmail,
		ColPhoneNumber,
		ColData,
		ColCreatedAt,
		ColUpdatedAt,
	}
)
