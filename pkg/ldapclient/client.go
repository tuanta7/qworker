package ldapclient

import (
	"crypto/tls"
	"github.com/go-ldap/ldap/v3"
	"net"
	neturl "net/url"
	"sync"
	"time"
)

type LDAPClient struct {
	lock        sync.Mutex
	pools       map[string]*LDAPPool
	skipVerify  bool
	maxPoolWait time.Duration
}

func NewLDAPClient(skipVerify bool) *LDAPClient {
	return &LDAPClient{
		lock:       sync.Mutex{},
		pools:      make(map[string]*LDAPPool),
		skipVerify: skipVerify,
	}
}

func (c *LDAPClient) NewLDAPPool(url string, maxConns int) (LDAPPool, error) {
	p := &ldapPool{
		lock:     sync.Mutex{},
		conns:    make(chan *ldap.Conn, maxConns),
		maxConns: maxConns,
		maxWait:  c.maxPoolWait,
	}

	for i := 0; i < p.maxConns; i++ {
		conn, err := c.NewConnection(url, 0)
		if err != nil {
			return nil, err
		}
		p.conns <- conn
	}

	return p, nil
}

func (c *LDAPClient) NewConnection(url string, timeout time.Duration) (*ldap.Conn, error) {
	dialerConfig := &net.Dialer{
		Timeout: timeout,
	}

	conn, err := ldap.DialURL(url, ldap.DialWithDialer(dialerConfig))
	if err != nil {
		return nil, err
	}

	serverName, _ := extractServerName(url)
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
		return "", err
	}

	return host, nil
}
