package utils

import "time"

func TimeToLDAPString(t time.Time) string {
	return t.Format("20060102150405.0Z")
}

func LDAPStringToTime(ldapTime string) (time.Time, error) {
	return time.Parse("20060102150405Z", ldapTime)
}
