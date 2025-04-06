package schema

import (
	"encoding/json"

	// Packages
	pg "github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type DatabaseName string

type Database struct {
	Name  string   `json:"name,omitempty" arg:"" help:"Name"`
	Owner string   `json:"owner,omitempty" help:"Owner"`
	Priv  []string `json:"priv,omitempty" help:"Access privileges"`
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

////////////////////////////////////////////////////////////////////////////////
// READER

func (d *Database) Scan(row pg.Row) error {
	return row.Scan(&d.Name, &d.Owner, &d.Priv)
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
	databaseGet  = databaseSelect + `WHERE "name" = @name`
	databaseList = `WITH q AS (` + databaseSelect + `) SELECT * FROM q ${where}`
)
