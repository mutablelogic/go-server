package schema

import (
	"encoding/json"
	"strings"

	// Packages
	pg "github.com/djthorpe/go-pg"
	types "github.com/djthorpe/go-pg/pkg/types"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type DatabaseName string

type Database struct {
	Oid uint32 `json:"oid"`
	DatabaseMeta
	Size uint64 `json:"bytes,omitempty" help:"Size of database in bytes"`
}

type DatabaseMeta struct {
	Name  string  `json:"name,omitempty" arg:"" help:"Name"`
	Owner string  `json:"owner,omitempty" help:"Owner"`
	Acl   ACLList `json:"acl,omitempty" help:"Access privileges"`
}

type DatabaseListRequest struct {
	pg.OffsetLimit
}

type DatabaseList struct {
	Count uint64     `json:"count"`
	Body  []Database `json:"body,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (d Database) String() string {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (d DatabaseMeta) String() string {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (d DatabaseListRequest) String() string {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (d DatabaseList) String() string {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// SELECT

func (d DatabaseListRequest) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Set empty where
	bind.Set("where", "")
	bind.Set("orderby", "ORDER BY name ASC")

	// Bind offset and limit
	d.OffsetLimit.Bind(bind, DatabaseListLimit)

	// Return query
	switch op {
	case pg.List:
		return databaseList, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("unsupported DatabaseListRequest operation %q", op)
	}
}

func (d DatabaseName) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Set name
	if name, err := d.name(); err != nil {
		return "", err
	} else {
		bind.Set("name", name)
	}

	// Set force
	if force, ok := bind.Get("force").(bool); ok && force {
		bind.Set("with", "WITH FORCE")
	} else {
		bind.Set("with", "")
	}

	// Return query
	switch op {
	case pg.Get:
		return databaseGet, nil
	case pg.Update:
		return databaseRename, nil
	case pg.Delete:
		return databaseDelete, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("unsupported DatabaseName operation %q", op)
	}
}

func (d DatabaseMeta) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Set name
	if name := strings.TrimSpace(d.Name); name == "" {
		return "", httpresponse.ErrBadRequest.With("name is missing")
	} else {
		bind.Set("name", name)
	}

	// Return query
	switch op {
	case pg.Update:
		return databaseUpdate, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("unsupported Database operation %q", op)
	}
}

////////////////////////////////////////////////////////////////////////////////
// WRITER

func (d DatabaseMeta) Insert(bind *pg.Bind) (string, error) {
	// Set name
	if name, err := d.name(); err != nil {
		return "", err
	} else {
		bind.Set("name", name)
	}

	// Set with
	bind.Set("with", d.with(true))

	// Return success
	return databaseCreate, nil
}

func (d DatabaseMeta) Update(bind *pg.Bind) error {
	// With
	bind.Set("with", d.with(false))

	// Return success
	return nil
}

func (d DatabaseName) Insert(bind *pg.Bind) (string, error) {
	return "", httpresponse.ErrNotImplemented.With("DatabaseName.Insert")
}

func (d DatabaseName) Update(bind *pg.Bind) error {
	// Set name
	if name, err := d.name(); err != nil {
		return err
	} else {
		bind.Set("old_name", name)
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// READER

func (d *Database) Scan(row pg.Row) error {
	var priv []string
	d.Acl = ACLList{}
	if err := row.Scan(&d.Oid, &d.Name, &d.Owner, &priv, &d.Size); err != nil {
		return err
	}
	for _, v := range priv {
		item, err := NewACLItem(v)
		if err != nil {
			return err
		}
		d.Acl.Append(item)
	}
	return nil
}

func (n *DatabaseList) Scan(row pg.Row) error {
	var database Database
	if err := database.Scan(row); err != nil {
		return err
	} else {
		n.Body = append(n.Body, database)
	}
	return nil
}

func (n *DatabaseList) ScanCount(row pg.Row) error {
	return row.Scan(&n.Count)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (d DatabaseMeta) name() (string, error) {
	return DatabaseName(d.Name).name()
}

func (d DatabaseName) name() (string, error) {
	if name := strings.TrimSpace(string(d)); name == "" {
		return "", httpresponse.ErrBadRequest.With("name is missing")
	} else if strings.HasPrefix(name, reservedPrefix) {
		return "", httpresponse.ErrBadRequest.Withf("name cannot be prefixed with %q", reservedPrefix)
	} else {
		return name, nil
	}
}

func (d DatabaseMeta) with(insert bool) string {
	var with []string
	if owner := strings.TrimSpace(d.Owner); owner != "" {
		if insert {
			with = append(with, "WITH OWNER "+types.DoubleQuote(d.Owner))
		} else {
			with = append(with, "OWNER TO "+types.DoubleQuote(d.Owner))
		}
	}

	// Return the with clause
	if len(with) > 0 {
		return strings.Join(with, " ")
	}
	return ""
}

////////////////////////////////////////////////////////////////////////////////
// SQL

const (
	databaseSelect = `
		WITH s AS (SELECT
			D.oid AS "oid", D.datname AS "name", R.rolname AS "owner", D.datacl AS "acl", pg_database_size(D.oid) AS "size"
		FROM
			${"schema"}."pg_database" D
		JOIN
			${"schema"}."pg_roles" R ON D.datdba = R.oid
		WHERE
			D.datistemplate = false) SELECT * FROM s
	`
	databaseGet    = databaseSelect + ` WHERE "name" = @name`
	databaseList   = `WITH q AS (` + databaseSelect + `) SELECT * FROM q ${where} ${orderby}`
	databaseCreate = `CREATE DATABASE ${"name"} ${with}`
	databaseDelete = `DROP DATABASE ${"name"} ${with}`
	databaseRename = `ALTER DATABASE ${"old_name"} RENAME TO ${"name"}`
	databaseUpdate = `
		ALTER DATABASE ${"name"} ${with}
	`
)
