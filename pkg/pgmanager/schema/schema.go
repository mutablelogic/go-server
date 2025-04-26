package schema

import (
	"encoding/json"
	"strings"

	// Packages
	pg "github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type SchemaName string

type SchemaMeta struct {
	Name  string  `json:"name,omitempty" arg:"" help:"Name"`
	Owner string  `json:"owner,omitempty" help:"Owner"`
	Acl   ACLList `json:"acl,omitempty" help:"Access privileges"`
}

type Schema struct {
	Oid      uint32 `json:"oid"`
	Database string `json:"database,omitempty" help:"Database"`
	SchemaMeta
	Size uint64 `json:"bytes,omitempty" help:"Size of schema in bytes"`
}

type SchemaListRequest struct {
	Database *string `json:"database,omitempty" help:"Database"`
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
	// Order
	bind.Set("orderby", `ORDER BY name ASC`)

	// Where
	bind.Del("where")
	if database := types.PtrString(d.Database); database != "" {
		bind.Append("where", `database = `+types.Quote(database))
	}
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

func (s SchemaName) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Set name
	if name := strings.TrimSpace(string(s)); name == "" {
		return "", httpresponse.ErrBadRequest.With("schema is missing")
	} else {
		bind.Set("name", name)
	}

	// Set force
	if force, ok := bind.Get("force").(bool); ok && force {
		bind.Set("with", "CASCADE")
	} else {
		bind.Set("with", "")
	}

	// Return query
	switch op {
	case pg.Get:
		return schemaGet, nil
	case pg.Update:
		return schemaRename, nil
	case pg.Delete:
		return schemaDelete, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("unsupported SchemaName operation %q", op)
	}
}

func (s SchemaMeta) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Set name
	if name := strings.TrimSpace(s.Name); name == "" {
		return "", httpresponse.ErrBadRequest.With("schema is missing")
	} else {
		bind.Set("name", name)
	}

	// Return query
	switch op {
	case pg.Update:
		return schemaUpdate, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("unsupported SchemaMeta operation %q", op)
	}
}

////////////////////////////////////////////////////////////////////////////////
// READER

func (s *Schema) Scan(row pg.Row) error {
	var priv []string
	s.Acl = ACLList{}
	if err := row.Scan(&s.Oid, &s.Database, &s.Name, &s.Owner, &priv, &s.Size); err != nil {
		return err
	}
	for _, v := range priv {
		item, err := NewACLItem(v)
		if err != nil {
			return err
		}
		s.Acl.Append(item)
	}
	return nil
}

func (s *SchemaList) Scan(row pg.Row) error {
	var schema Schema
	if err := schema.Scan(row); err != nil {
		return err
	} else {
		s.Body = append(s.Body, schema)
	}
	return nil
}

func (s *SchemaList) ScanCount(row pg.Row) error {
	return row.Scan(&s.Count)
}

////////////////////////////////////////////////////////////////////////////////
// WRITER

func (s SchemaMeta) Insert(bind *pg.Bind) (string, error) {
	// Set name
	if name := strings.TrimSpace(s.Name); name == "" {
		return "", httpresponse.ErrBadRequest.With("name is missing")
	} else if strings.HasPrefix(name, reservedPrefix) {
		return "", httpresponse.ErrBadRequest.Withf("cannot create a schema prefixed with %q", reservedPrefix)
	} else {
		bind.Set("name", name)
	}

	// Set with
	if with, err := s.with(true); err != nil {
		return "", err
	} else {
		bind.Set("with", with)
	}

	// Return success
	return schemaCreate, nil
}

func (s SchemaMeta) Update(bind *pg.Bind) error {
	// With
	if with, err := s.with(false); err != nil {
		return err
	} else {
		bind.Set("with", with)
	}

	// Return success
	return nil
}

func (d SchemaName) Insert(bind *pg.Bind) (string, error) {
	return "", httpresponse.ErrNotImplemented.With("SchemaName.Insert")
}

func (d SchemaName) Update(bind *pg.Bind) error {
	if name := strings.TrimSpace(string(d)); name == "" {
		return httpresponse.ErrBadRequest.With("name is missing")
	} else if strings.HasPrefix(name, reservedPrefix) {
		return httpresponse.ErrBadRequest.Withf("cannot create a schema prefixed with %q", reservedPrefix)
	} else if strings.ToLower(name) == defaultSchema {
		return httpresponse.ErrBadRequest.Withf("cannot rename schema %q", name)
	} else {
		bind.Set("old_name", name)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (s SchemaMeta) with(insert bool) (string, error) {
	var with []string
	if owner := strings.TrimSpace(s.Owner); owner != "" {
		if insert {
			with = append(with, "AUTHORIZATION "+types.DoubleQuote(s.Owner))
		} else {
			with = append(with, "OWNER TO "+types.DoubleQuote(s.Owner))
		}
	}

	// Return the with clause
	if len(with) > 0 {
		return strings.Join(with, " "), nil
	}

	if insert {
		return "", httpresponse.ErrBadRequest.With("missing owner")
	} else {
		return "", nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// SQL

const (
	SchemaDef    = `schema ("oid" OID, "database" TEXT, "name" TEXT, "owner" TEXT, "acl" TEXT[], "size" BIGINT)`
	schemaSelect = `
		WITH sc AS (
			SELECT
				S.oid AS "oid", current_database() AS "database", S.nspname AS "name", R.rolname AS "owner", S.nspacl AS "acl", COALESCE(SUM(pg_relation_size(C.oid)),0) AS "size"
			FROM
				"pg_catalog"."pg_namespace" S
			LEFT JOIN
				"pg_catalog"."pg_roles" R ON S.nspowner = R.oid
			LEFT JOIN
				"pg_catalog"."pg_class" C ON C.relnamespace = S.oid
			WHERE
				S.nspname NOT LIKE 'pg_%' AND S.nspname != 'information_schema'
			GROUP BY
				1, 2, 3, 4, 5				
		) SELECT * FROM sc`
	schemaGet    = schemaSelect + ` WHERE "name" = ${'name'}`
	schemaList   = `WITH q AS (` + schemaSelect + `) SELECT * FROM q ${where} ${orderby}`
	schemaDelete = `DROP SCHEMA ${"name"} ${with}`
	schemaCreate = `CREATE SCHEMA ${"name"} ${with}`
	schemaRename = `ALTER SCHEMA ${"old_name"} RENAME TO ${"name"}`
	schemaUpdate = `ALTER SCHEMA ${"name"} ${with}`
)
