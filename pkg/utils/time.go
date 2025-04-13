package utils

import "time"

const (
	ldapTimeFormat = "20060102150405.0Z"
)

func TimeToLDAPString(t time.Time) string {
	return t.Format(ldapTimeFormat)
}

func LDAPStringToTime(ldapTime string) (time.Time, error) {
	return time.Parse(ldapTimeFormat, ldapTime)
}
