package ldapclient

import (
	"errors"
	"github.com/go-ldap/ldap/v3"
	"sync"
	"time"
)

type LDAPConn interface {
	Bind(username, password string) error
	UnauthenticatedBind(username string) error
	Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error)
	Close() (err error)
}

type LDAPPool interface {
	GetConn() (LDAPConn, error)
	ReturnConn(c LDAPConn)
	Close()
}

type ldapPool struct {
	lock     sync.Mutex
	maxConns int
	maxWait  time.Duration
	conns    chan LDAPConn
	closed   bool
}

func (p *ldapPool) GetConn() (LDAPConn, error) {
	select {
	case conn := <-p.conns:
		return conn, nil
	case <-time.After(p.maxWait):
		return nil, errors.New("pool timeout")
	}
}

func (p *ldapPool) ReturnConn(c LDAPConn) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.closed {
		return
	}

	if len(p.conns) < p.maxConns {
		_ = c.UnauthenticatedBind("")
		p.conns <- c
	}
}

func (p *ldapPool) Close() {
	p.lock.Lock()
	if !p.closed {
		p.closed = true
		close(p.conns)
	}

	for conn := range p.conns {
		conn.Close()
	}

	p.lock.Unlock()
}
