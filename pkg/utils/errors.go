package utils

import "errors"

var (
	ErrConnectorNotFound = errors.New("connector not found")
	ErrJobNotFound       = errors.New("job not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrNoUserProvided    = errors.New("no users provided")
	ErrTaskConflict      = errors.New("task conflict")
)
