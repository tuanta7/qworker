package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMapperScan(t *testing.T) {
	m := &Mapper{}

	t.Run("empty_mapper", func(t *testing.T) {
		err := m.Scan("")
		assert.Equal(t, nil, err)
		assert.Equal(t, Mapper{}, *m)
	})

	t.Run("nil_mapper", func(t *testing.T) {
		err := m.Scan(nil)
		assert.Equal(t, nil, err)
		assert.Equal(t, Mapper{}, *m)
	})

	t.Run("invalid_json", func(t *testing.T) {
		err := m.Scan("{")
		assert.Equal(t, nil, err)
		assert.Equal(t, Mapper{}, *m)
	})

	t.Run("success", func(t *testing.T) {
		mockMapper := &Mapper{
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

		s, err := mockMapper.Value()
		assert.Equal(t, nil, err)

		err = m.Scan(s)
		assert.Equal(t, nil, err)
		assert.Equal(t, *mockMapper, *m)
		assert.Equal(t, mockMapper.Email, m.Email)
		assert.Equal(t, "sn", m.Custom["lastName"])
	})
}
