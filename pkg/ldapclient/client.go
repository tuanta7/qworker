package ldapclient

import (
	"crypto/tls"
	"github.com/go-ldap/ldap/v3"
	"net"
	neturl "net/url"
	"sync"
	"time"
)

type LDAPClient interface {
	NewConnection(url string, timeout time.Duration) (LDAPConn, error)
	NewLDAPPool(url string, maxConns int) (LDAPPool, error)
}

type ldapClient struct {
	lock        sync.Mutex
	pools       map[string]LDAPPool
	skipVerify  bool
	maxPoolWait time.Duration
}

func NewLDAPClient(skipVerify bool) LDAPClient {
	return &ldapClient{
		lock:       sync.Mutex{},
		pools:      make(map[string]LDAPPool),
		skipVerify: skipVerify,
	}
}

func (c *ldapClient) NewLDAPPool(addr string, maxConns int) (LDAPPool, error) {
	p := &ldapPool{
		lock:     sync.Mutex{},
		conns:    make(chan LDAPConn, maxConns),
		maxConns: maxConns,
		maxWait:  c.maxPoolWait,
	}

	for i := 0; i < p.maxConns; i++ {
		conn, err := c.NewConnection(addr, 0)
		if err != nil {
			p.Close()
			return nil, err
		}
		p.conns <- conn
	}

	return p, nil
}

func (c *ldapClient) NewConnection(addr string, timeout time.Duration) (LDAPConn, error) {
	conn, err := ldap.DialURL(addr, ldap.DialWithDialer(&net.Dialer{
		Timeout: timeout,
	}))
	if err != nil {
		return nil, err
	}

	serverName, _ := extractServerName(addr)
	err = conn.StartTLS(&tls.Config{
		ServerName:         serverName,
		InsecureSkipVerify: c.skipVerify,
	})
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	return conn, nil
}

func extractServerName(rawURL string) (string, error) {
	url, err := neturl.Parse(rawURL)
	if err != nil {
		return "", err
	}

	host, _, err := net.SplitHostPort(url.Host)
	if err != nil {
		if err.Error() == "missing port in address" {
			return url.Host, nil
		}
		return "", err
	}

	return host, nil
}
