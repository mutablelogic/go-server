package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	// Packages
	pg "github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type RoleName string

type RoleMeta struct {
	Name                   string     `json:"name,omitempty" arg:"" help:"Role name"`
	Superuser              *bool      `json:"super,omitempty" help:"Superuser permission"`
	Inherit                *bool      `json:"inherit,omitempty" help:"Inherit permissions"`
	CreateRoles            *bool      `json:"createrole,omitempty" help:"Create roles permission"`
	CreateDatabases        *bool      `json:"createdb,omitempty" help:"Create databases permission"`
	Replication            *bool      `json:"replication,omitempty" help:"Replication permission"`
	ConnectionLimit        *uint64    `json:"conlimit,omitempty" help:"Connection limit"`
	BypassRowLevelSecurity *bool      `json:"bypassrls,omitempty" help:"Bypass row-level security"`
	Login                  *bool      `json:"login,omitempty" help:"Login permission"`
	Password               *string    `json:"password,omitempty" help:"Password"`
	Expires                *time.Time `json:"expires,omitzero" help:"Password expiration"`
	Groups                 []string   `json:"memberof,omitempty" help:"Group memberships"`
}

type Role struct {
	Oid uint32 `json:"oid"`
	RoleMeta
}

type RoleListRequest struct {
	pg.OffsetLimit
}

type RoleList struct {
	Count uint64 `json:"count"`
	Body  []Role `json:"body,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r Role) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (r RoleMeta) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (r RoleList) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (r RoleListRequest) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func RevokeGroupMembership(ctx context.Context, conn pg.Conn, group, member string) error {
	return conn.Exec(ctx, fmt.Sprintf("REVOKE %s FROM %s", group, member))
}

func GrantGroupMembership(ctx context.Context, conn pg.Conn, group, member string) error {
	return conn.Exec(ctx, fmt.Sprintf("GRANT %s TO %s", group, member))
}

////////////////////////////////////////////////////////////////////////////////
// SELECT

func (r RoleListRequest) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Set empty where
	bind.Set("where", "")

	// Bind offset and limit
	r.OffsetLimit.Bind(bind, RoleListLimit)

	// Return query
	switch op {
	case pg.List:
		return roleList, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("unsupported RoleListRequest operation %q", op)
	}
}

func (r RoleName) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Set name
	if name := strings.TrimSpace(string(r)); name == "" {
		return "", httpresponse.ErrBadRequest.With("name is missing")
	} else {
		bind.Set("name", name)
	}

	// Return query
	switch op {
	case pg.Get:
		return roleGet, nil
	case pg.Update:
		return roleRename, nil
	case pg.Delete:
		return roleDelete, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("unsupported RoleName operation %q", op)
	}
}

func (r RoleMeta) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Set name
	if name := strings.TrimSpace(r.Name); name == "" {
		return "", httpresponse.ErrBadRequest.With("name is missing")
	} else {
		bind.Set("name", name)
	}

	// Return query
	switch op {
	case pg.Update:
		return roleUpdate, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("unsupported RoleMeta operation %q", op)
	}
}

////////////////////////////////////////////////////////////////////////////////
// READER

func (r *Role) Scan(row pg.Row) error {
	var connlimit int64
	if err := row.Scan(&r.Oid, &r.Name, &r.Superuser, &r.Inherit, &r.CreateRoles, &r.CreateDatabases, &r.Replication, &connlimit, &r.BypassRowLevelSecurity, &r.Login, &r.Password, &r.Expires, &r.Groups); err != nil {
		return err
	}
	if connlimit >= 0 {
		r.ConnectionLimit = types.Uint64Ptr(uint64(connlimit))
	} else {
		r.ConnectionLimit = nil
	}
	if len(r.Groups) == 0 {
		r.Groups = nil
	}
	return nil
}

func (n *RoleList) Scan(row pg.Row) error {
	var role Role
	if err := role.Scan(row); err != nil {
		return err
	} else {
		n.Body = append(n.Body, role)
	}
	return nil
}

func (n *RoleList) ScanCount(row pg.Row) error {
	return row.Scan(&n.Count)
}

////////////////////////////////////////////////////////////////////////////////
// WRITER

func (r RoleMeta) Insert(bind *pg.Bind) (string, error) {
	// Name
	if name := strings.TrimSpace(r.Name); strings.HasPrefix(name, "pg_") {
		return "", httpresponse.ErrBadRequest.With("name cannot start with pg_")
	} else if !types.IsIdentifier(name) {
		return "", httpresponse.ErrBadRequest.With("name is invalid")
	} else {
		bind.Set("name", name)
	}

	// With
	bind.Set("with", r.with(true))

	// Return the query
	return roleCreate, nil
}

func (r RoleName) Insert(bind *pg.Bind) (string, error) {
	return "", httpresponse.ErrNotImplemented.With("RoleName.Insert")
}

func (r RoleName) Update(bind *pg.Bind) error {
	if name := strings.TrimSpace(string(r)); strings.HasPrefix(name, "pg_") {
		return httpresponse.ErrBadRequest.With("name cannot start with pg_")
	} else {
		bind.Set("old_name", name)
	}
	return nil
}

func (r RoleMeta) Update(bind *pg.Bind) error {
	// With
	bind.Set("with", r.with(false))

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (r RoleMeta) with(insert bool) string {
	var with []string
	opt := func(v string, b bool) string {
		if b {
			return v
		}
		return "NO" + v
	}
	if r.Superuser != nil {
		with = append(with, opt("SUPERUSER", types.PtrBool(r.Superuser)))
	}
	if r.CreateDatabases != nil {
		with = append(with, opt("CREATEDB", types.PtrBool(r.CreateDatabases)))
	}
	if r.CreateRoles != nil {
		with = append(with, opt("CREATEROLE", types.PtrBool(r.CreateRoles)))
	}
	if r.Replication != nil {
		with = append(with, opt("REPLICATION", types.PtrBool(r.Replication)))
	}
	if r.Inherit != nil {
		with = append(with, opt("INHERIT", types.PtrBool(r.Inherit)))
	}
	if r.Login != nil {
		with = append(with, opt("LOGIN", types.PtrBool(r.Login)))
	}
	if r.BypassRowLevelSecurity != nil {
		with = append(with, opt("BYPASSRLS", types.PtrBool(r.BypassRowLevelSecurity)))
	}
	if r.ConnectionLimit != nil {
		with = append(with, fmt.Sprintf("CONNECTION LIMIT %v", types.PtrUint64(r.ConnectionLimit)))
	}
	if r.Password != nil {
		if password := types.PtrString(r.Password); password == pgObfuscatedPassword {
			// Do nothing
		} else if password == "" {
			with = append(with, "PASSWORD NULL")
		} else {
			with = append(with, fmt.Sprintf("PASSWORD %v", types.Quote(password)))
		}
	}
	if expires := types.PtrTime(r.Expires).UTC(); !expires.IsZero() {
		with = append(with, fmt.Sprintf("VALID UNTIL %v", types.Quote(expires.Format(pgTimestampFormat))))
	}
	if len(r.Groups) > 0 && insert {
		with = append(with, "IN ROLE "+strings.Join(r.Groups, ", "))
	}

	// Return the with clause
	if len(with) > 0 {
		return "WITH " + strings.Join(with, " ")
	}
	return ""
}

////////////////////////////////////////////////////////////////////////////////
// SQL

const (
	roleCreate = `
		CREATE ROLE ${"name"} ${with}
	`
	roleRename = `
		ALTER ROLE ${"old_name"} RENAME TO ${"name"}
	`
	roleDelete = `
		DROP ROLE ${"name"}
	`
	roleUpdate = `
		ALTER ROLE ${"name"} ${with}
	`
	roleSelect = `
		WITH roles AS (
			SELECT
				"oid", "rolname", "rolsuper", "rolinherit", "rolcreaterole", "rolcreatedb", "rolreplication", "rolconnlimit", "rolbypassrls", "rolcanlogin", "rolpassword", "rolvaliduntil",
                ARRAY(SELECT R2.rolname FROM "pg_catalog".pg_auth_members M JOIN "pg_catalog".pg_roles R2 ON M.roleid = R2.oid WHERE M.member = R.oid) AS groups
			FROM
				${"schema"}."pg_roles" R
			WHERE
				"rolname" NOT LIKE 'pg_%'
		) SELECT * FROM roles
	`
	roleGet  = roleSelect + `WHERE rolname = @name`
	roleList = `WITH q AS (` + roleSelect + `) SELECT * FROM q ${where}`
)
