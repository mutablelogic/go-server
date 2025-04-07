package schema

import (
	"encoding/json"

	pg "github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type SchemaName string

type SchemaMeta struct {
	Name  string     `json:"name,omitempty" arg:"" help:"Name"`
	Owner string     `json:"owner,omitempty" help:"Owner"`
	Acl   []*ACLItem `json:"acl,omitempty" help:"Access privileges"`
}

type Schema struct {
	Database string `json:"database,omitempty" help:"Database"`
	SchemaMeta
}

type SchemaListRequest struct {
	Database string `json:"database,omitempty" help:"Database"`
	pg.OffsetLimit
}

type SchemaList struct {
	Count uint64   `json:"count"`
	Body  []Schema `json:"body,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s SchemaMeta) String() string {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (s Schema) String() string {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (s SchemaListRequest) String() string {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (s SchemaList) String() string {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// SELECT

func (d SchemaListRequest) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Set empty where
	bind.Del("where")

	// Database
	if d.Database != "" {
		bind.Append("where", `database = `+bind.Set("database", d.Database))
	}

	// Set where
	if where := bind.Join("where", " AND "); where != "" {
		bind.Set("where", `WHERE `+where)
	} else {
		bind.Set("where", "")
	}

	// Bind offset and limit
	d.OffsetLimit.Bind(bind, SchemaListLimit)

	// Return query
	switch op {
	case pg.List:
		return schemaList, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("unsupported SchemaListRequest operation %q", op)
	}
}

////////////////////////////////////////////////////////////////////////////////
// SQL

const (
	schemaSelect = `
		WITH db AS (
			SELECT
				S.nspname AS "name", R.rolname AS "owner", S.datacl AS "acl"
			FROM
				${"schema"}."pg_namespace" S
			JOIN
				${"schema"}."pg_roles" R
			ON 
				S.nspowner = R.oid
		) SELECT * FROM db`
	schemaGet  = schemaSelect + ` WHERE "name" = @name`
	schemaList = `WITH q AS (` + schemaSelect + `) SELECT * FROM q ${where}`
)
