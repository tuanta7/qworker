package domain

const (
	ColData      = "data"
	ColCreatedAt = "created_at"
	ColUpdatedAt = "updated_at"

	TableConnector   = "private.connector"
	ColConnectorID   = "id"
	ColConnectorType = "connector_type"
	ColDisplayName   = "display_name"
	ColEnabled       = "enabled"
	ColLastSync      = "last_sync"

	TableUser        string = "private.user"
	ColUserID        string = "id"
	ColUsername      string = "username"
	ColFullName      string = "full_name"
	ColPhoneNumber   string = "phone_number"
	ColEmail         string = "email"
	ColEmailVerified string = "email_verified"
	ColActive        string = "active"
	ColSourceID      string = "source_id"
)

var (
	AllConnectorCols = []string{
		ColConnectorID,
		ColConnectorType,
		ColDisplayName,
		ColEnabled,
		ColLastSync,
		ColData,
		ColCreatedAt,
		ColUpdatedAt,
	}

	AllUserCols = []string{
		ColUserID,
		ColUsername,
		ColFullName,
		ColPhoneNumber,
		ColEmail,
		ColEmailVerified,
		ColActive,
		ColSourceID,
		ColData,
		ColCreatedAt,
		ColUpdatedAt,
	}

	AllUserSyncCols = []string{
		ColUserID,
		ColUsername,
		ColFullName,
		ColPhoneNumber,
		ColEmail,
		ColSourceID,
		ColData,
		ColCreatedAt,
		ColUpdatedAt,
	}
)
