package ldapclient

import (
	"github.com/go-ldap/ldap/v3"
)

type SearchOptions func(*ldap.SearchRequest)

type LDAPConnection interface {
	Bind(username, password string) error
	Search(baseDN string, filter map[string]any, opts ...SearchOptions) []*ldap.Entry
	Close() error
}

type ldapConnection struct {
	*ldap.Conn
}

func (c *ldapConnection) Bind(username, password string) error {
	return c.Conn.Bind(username, password)
}

func (c *ldapConnection) Search(baseDN string, filter map[string]any, opts ...SearchOptions) []*ldap.Entry {
	return nil
}

func (c *ldapConnection) Close() error {
	return c.Conn.Close()
}

func WithPagination(page uint64, pageSize uint64) SearchOptions {
	return func(r *ldap.SearchRequest) {
		r.Controls = nil
	}
}
