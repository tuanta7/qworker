package ldapclient

import (
	"crypto/tls"
	"github.com/go-ldap/ldap/v3"
	"net"
	"sync"
	"time"
)

type LDAPClient struct {
	lock       sync.Mutex
	pools      map[string]*LDAPPool
	skipVerify bool
}

func (c *LDAPClient) NewLDAPPool(url string, maxConns int) (LDAPPool, error) {
	if maxConns < 1 {
		maxConns = 1
	}

	p := &ldapPool{
		conns:    make(chan LDAPConnection, maxConns),
		maxConns: maxConns,
		url:      url,
	}

	return p, nil
}

func (c *LDAPClient) NewConnection(url string, timeout time.Duration) (LDAPConnection, error) {
	if timeout < 0 {
		timeout = 10 * time.Second
	}

	dialerConfig := &net.Dialer{
		Timeout: timeout,
	}

	conn, err := ldap.DialURL(url, ldap.DialWithDialer(dialerConfig))
	if err != nil {
		return nil, err
	}

	err = conn.StartTLS(&tls.Config{
		InsecureSkipVerify: c.skipVerify,
	})
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	return &ldapConnection{Conn: conn}, nil
}
