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
	Name  string     `json:"name,omitempty" arg:"" help:"Name"`
	Owner string     `json:"owner,omitempty" help:"Owner"`
	Acl   []*ACLItem `json:"acl,omitempty" help:"Access privileges"`
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
	if name := strings.TrimSpace(string(d)); name == "" {
		return "", httpresponse.ErrBadRequest.With("name is missing")
	} else if strings.HasPrefix(name, reservedPrefix) && (op == pg.Update || op == pg.Delete) {
		return "", httpresponse.ErrBadRequest.Withf("cannot alter a system database %q", name)
	} else {
		bind.Set("name", name)
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

func (d Database) Select(bind *pg.Bind, op pg.Op) (string, error) {
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

func (d Database) Insert(bind *pg.Bind) (string, error) {
	// Set name
	if name := strings.TrimSpace(d.Name); name == "" {
		return "", httpresponse.ErrBadRequest.With("name is missing")
	} else if strings.HasPrefix(name, reservedPrefix) {
		return "", httpresponse.ErrBadRequest.Withf("cannot create a database prefixed with %q", reservedPrefix)
	} else {
		bind.Set("name", name)
	}

	// Set with
	bind.Set("with", d.with(true))

	// Return success
	return databaseCreate, nil
}

func (d Database) Update(bind *pg.Bind) error {
	// With
	bind.Set("with", d.with(false))

	// Return success
	return nil
}

func (d DatabaseName) Insert(bind *pg.Bind) (string, error) {
	return "", httpresponse.ErrNotImplemented.With("DatabaseName.Insert")
}

func (d DatabaseName) Update(bind *pg.Bind) error {
	if name := strings.TrimSpace(string(d)); name == "" {
		return httpresponse.ErrBadRequest.With("name is missing")
	} else {
		bind.Set("old_name", name)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// READER

func (d *Database) Scan(row pg.Row) error {
	var priv []string
	if err := row.Scan(&d.Name, &d.Owner, &priv); err != nil {
		return err
	}
	for _, v := range priv {
		item, err := NewACLItem(v)
		if err != nil {
			return err
		}
		d.Acl = append(d.Acl, item)
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

func (d Database) with(insert bool) string {
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
		WITH db AS (
			SELECT
				D.datname AS "name", R.rolname AS "owner", D.datacl AS "acl"
			FROM
				${"schema"}."pg_database" D
			JOIN
				${"schema"}."pg_roles" R
			ON 
				D.datdba = R.oid
			WHERE
				D.datistemplate = false
		) SELECT * FROM db`
	databaseGet    = databaseSelect + ` WHERE "name" = @name`
	databaseList   = `WITH q AS (` + databaseSelect + `) SELECT * FROM q ${where}`
	databaseCreate = `CREATE DATABASE ${"name"} ${with}`
	databaseDelete = `DROP DATABASE ${"name"}`
	databaseRename = `ALTER DATABASE ${"old_name"} RENAME TO ${"name"}`
	databaseUpdate = `
		ALTER DATABASE ${"name"} ${with}
	`
)
