package schema

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	CatalogSchema = "pg_catalog"
)

const (
	// Maximum number of roles/databases/schemas/objects/connections to return in a list query
	RoleListLimit       = 100
	DatabaseListLimit   = 100
	SchemaListLimit     = 100
	ObjectListLimit     = 100
	ConnectionListLimit = 100
)

const (
	pgTimestampFormat    = "2006-01-02 15:04:05"
	pgObfuscatedPassword = "********"
	defaultAclRole       = "PUBLIC"
	defaultSchema        = "public"
	schemaSeparator      = '/'
	reservedPrefix       = "pg_"
)
