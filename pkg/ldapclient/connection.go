package ldapclient

import (
	"github.com/go-ldap/ldap/v3"
	"time"
)

type SearchOptions func(*ldap.SearchRequest)

type LDAPConnection interface {
	Bind(username, password string) error
	Search(baseDN string, filter string, timeout time.Duration, opts ...SearchOptions) (*ldap.SearchResult, error)
	Close() error
}

type ldapConnection struct {
	*ldap.Conn
}

func (c *ldapConnection) Bind(username, password string) error {
	return c.Conn.Bind(username, password)
}

func (c *ldapConnection) Search(baseDN string, filter string, timeout time.Duration, opts ...SearchOptions) (*ldap.SearchResult, error) {
	searchRequest := &ldap.SearchRequest{
		BaseDN:       baseDN,
		Filter:       filter,
		DerefAliases: ldap.NeverDerefAliases,
		SizeLimit:    0,
		TimeLimit:    int(timeout / time.Second),
	}

	for _, opt := range opts {
		opt(searchRequest)
	}

	return c.Conn.Search(searchRequest)
}

func (c *ldapConnection) Close() error {
	return c.Conn.Close()
}

func WithScope(scope int) SearchOptions {
	return func(r *ldap.SearchRequest) {
		_, exist := ldap.ScopeMap[scope]
		if exist {
			r.Scope = scope
		}
	}
}

func WithPagination(pagingControl *ldap.ControlPaging) SearchOptions {
	return func(r *ldap.SearchRequest) {
		r.Controls = append(r.Controls, pagingControl)
	}
}
