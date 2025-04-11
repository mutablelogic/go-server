package schema

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	CatalogSchema = "pg_catalog"
)

const (
	// Maximum number of roles/databases to return in a list query
	RoleListLimit     = 100
	DatabaseListLimit = 100
	SchemaListLimit   = 100
	ObjectListLimit   = 100
)

const (
	pgTimestampFormat    = "2006-01-02 15:04:05"
	pgObfuscatedPassword = "********"
	defaultAclRole       = "PUBLIC"
	schemaSeparator      = '/'
	reservedPrefix       = "pg_"
)
