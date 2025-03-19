package domain

import "time"

type LdapConnector struct {
	URL                   string        `json:"url"`
	ConnectTimeout        time.Duration `json:"connectTimeout"`
	ReadTimeout           time.Duration `json:"readTimeout"`
	SystemAccountDN       string        `json:"systemAccountDn"`
	SystemAccountPassword string        `json:"systemAccountPassword"`
	UsernameAttribute     string        `json:"usernameAttribute"`
	BaseDN                string        `json:"baseDn"`
	SyncSettings          SyncSettings  `json:"syncSettings"`
}

type SyncSettings struct {
	BatchSize     uint32        `json:"batchSize"`
	IncSync       bool          `json:"incrementalSyncEnabled"`
	IncSyncPeriod time.Duration `json:"incrementalSyncPeriod"`
}
