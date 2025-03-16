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
	lock       sync.Mutex
	pools      map[string]*LDAPPool
	skipVerify bool
}

func NewLDAPClient(skipVerify bool) *LDAPClient {
	return &LDAPClient{
		lock:       sync.Mutex{},
		pools:      make(map[string]*LDAPPool),
		skipVerify: skipVerify,
	}
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

	serverName, _ := extractServerName(url)
	err = conn.StartTLS(&tls.Config{
		ServerName:         serverName,
		InsecureSkipVerify: c.skipVerify,
	})
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	return &ldapConnection{Conn: conn}, nil
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
