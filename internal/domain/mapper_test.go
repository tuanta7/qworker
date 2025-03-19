package domain

import (
	testifyassert "github.com/stretchr/testify/assert"
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
	assert := testifyassert.New(t)

	m := &Mapper{}

	err := m.Scan("")
	assert.Equal(nil, err)
	assert.Equal(Mapper{}, *m)

	err = m.Scan(nil)
	assert.Equal(nil, err)
	assert.Equal(Mapper{}, *m)

	err = m.Scan("{")
	assert.Equal(nil, err)
	assert.Equal(Mapper{}, *m)

	s, err := mockMapper.Value()
	assert.Equal(nil, err)

	err = m.Scan(s)
	assert.Equal(nil, err)
	assert.Equal(*mockMapper, *m)
	assert.Equal(mockMapper.Email, m.Email)
	assert.Equal("sn", m.Custom["lastName"])
}
