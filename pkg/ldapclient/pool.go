package ldapclient

import (
	"errors"
	"github.com/go-ldap/ldap/v3"
	"sync"
	"time"
)

type LDAPPool interface {
	Get() (*ldap.Conn, error)
	Return(c *ldap.Conn)
	Close()
}

type ldapPool struct {
	lock     sync.Mutex
	maxConns int
	maxWait  time.Duration
	conns    chan *ldap.Conn
	closed   bool
}

func (p *ldapPool) Get() (*ldap.Conn, error) {
	select {
	case conn := <-p.conns:
		return conn, nil
	case <-time.After(p.maxWait):
		return nil, errors.New("pool timeout")
	}
}

func (p *ldapPool) Return(c *ldap.Conn) {
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
	p.lock.Unlock()
}
