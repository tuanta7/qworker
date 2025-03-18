package ldapclient

import (
	"errors"
	"sync"
	"time"
)

type LDAPPool interface {
	Get() (LDAPConnection, error)
	Return(c LDAPConnection)
	Close()
}

type ldapPool struct {
	lock     sync.Mutex //  locks actions for a specific struct instance
	maxConns int
	maxWait  time.Duration
	conns    chan LDAPConnection
	closed   bool
}

func (p *ldapPool) Get() (LDAPConnection, error) {
	select {
	case conn := <-p.conns:
		return conn, nil
	case <-time.After(p.maxWait):
		return nil, errors.New("pool timeout")
	}
}

func (p *ldapPool) Return(c LDAPConnection) {
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
