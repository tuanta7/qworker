package ldapclient

import "sync"

type LDAPPool interface {
	Get() (LDAPConnection, error)
	Return()
	Close()
}

type ldapPool struct {
	lock     sync.Mutex
	conns    chan LDAPConnection
	maxConns int
	url      string
}

func (p *ldapPool) Get() (LDAPConnection, error) {
	return nil, nil
}

func (p *ldapPool) Return() {}

func (p *ldapPool) Close() {}
