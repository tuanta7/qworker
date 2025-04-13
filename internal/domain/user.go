package domain

import (
	"time"
)

type User struct {
	UserID        string    `json:"id"`
	Username      string    `json:"username"`
	FullName      string    `json:"fullName"`
	PhoneNumber   string    `json:"phoneNumber"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"emailVerified"`
	Active        bool      `json:"active"`
	SourceID      *uint64   `json:"sourceID"`
	Data          TextData  `json:"data"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
