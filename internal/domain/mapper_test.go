package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var mockMapper = &Mapper{
	ExternalID:  "uid",
	FullName:    "cn",
	Email:       "mail",
	PhoneNumber: "mobile",
	CreatedAt:   "whenCreated",
	UpdatedAt:   "whenChanges",
	Custom: map[string]string{
		"lastName": "sn",
		"username": "sAMAccountName", // override
	},
}

func TestMapper_Scan(t *testing.T) {
	m := &Mapper{}

	err := m.Scan("")
	assert.Equal(t, nil, err)
	assert.Equal(t, Mapper{}, *m)

	err = m.Scan(nil)
	assert.Equal(t, nil, err)
	assert.Equal(t, Mapper{}, *m)

	err = m.Scan("{")
	assert.Equal(t, nil, err)
	assert.Equal(t, Mapper{}, *m)

	s, err := mockMapper.Value()
	assert.Equal(t, nil, err)

	err = m.Scan(s)
	assert.Equal(t, nil, err)
	assert.Equal(t, *mockMapper, *m)
	assert.Equal(t, mockMapper.Email, m.Email)
	assert.Equal(t, "sn", m.Custom["lastName"])
}
