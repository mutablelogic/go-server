package schema

import (
	"encoding/json"
	"strings"

	pg "github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ObjectName string

type ObjectMeta struct {
	Name  string     `json:"name,omitempty" arg:"" help:"Name"`
	Owner string     `json:"owner,omitempty" help:"Owner"`
	Acl   []*ACLItem `json:"acl,omitempty" help:"Access privileges"`
}

type Object struct {
	Oid      uint32 `json:"oid"`
	Database string `json:"database,omitempty" help:"Database"`
	Schema   string `json:"schema,omitempty" help:"Schema"`
	Type     string `json:"type,omitempty" help:"Type"`
	ObjectMeta
}

type ObjectListRequest struct {
	Database *string `json:"database,omitempty" help:"Database"`
	Schema   *string `json:"schema,omitempty" help:"Schema"`
	pg.OffsetLimit
}

type ObjectList struct {
	Count uint64   `json:"count"`
	Body  []Object `json:"body,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (o ObjectMeta) String() string {
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (o Object) String() string {
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (o ObjectListRequest) String() string {
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (o ObjectList) String() string {
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// SELECT

func (o ObjectListRequest) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Order
	bind.Set("orderby", `ORDER BY schema ASC, name ASC`)

	// Where
	bind.Del("where")
	if schema := strings.TrimSpace(types.PtrString(o.Schema)); schema != "" {
		bind.Append("where", `schema = `+bind.Set("schema", schema))
	}
	if where := bind.Join("where", " AND "); where != "" {
		bind.Set("where", `WHERE `+where)
	} else {
		bind.Set("where", "")
	}

	// Bind offset and limit
	o.OffsetLimit.Bind(bind, ObjectListLimit)

	// Return query
	switch op {
	case pg.List:
		return objectList, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("unsupported ObjectListRequest operation %q", op)
	}
}

////////////////////////////////////////////////////////////////////////////////
// READER

func (o *Object) Scan(row pg.Row) error {
	var priv []string
	if err := row.Scan(&o.Oid, &o.Database, &o.Schema, &o.Name, &o.Type, &o.Owner, &priv); err != nil {
		return err
	}
	for _, v := range priv {
		item, err := NewACLItem(v)
		if err != nil {
			return err
		}
		o.Acl = append(o.Acl, item)
	}
	return nil
}

func (o *ObjectList) Scan(row pg.Row) error {
	var object Object
	if err := object.Scan(row); err != nil {
		return err
	} else {
		o.Body = append(o.Body, object)
	}
	return nil
}

func (o *ObjectList) ScanCount(row pg.Row) error {
	return row.Scan(&o.Count)
}

////////////////////////////////////////////////////////////////////////////////
// SQL

const (
	ObjectDef    = `object ("oid" OID, "database" TEXT, "schema" TEXT, "object" TEXT, "kind" TEXT, "owner" TEXT, "acl" TEXT[])`
	objectSelect = `
		WITH objects AS (
			SELECT
				C.oid AS oid,
				current_database() AS database,
				N.nspname AS schema,
				C.relname AS name,
				CASE C.relkind
					WHEN 'r' THEN 'TABLE'
					WHEN 'v' THEN 'VIEW'
					WHEN 'i' THEN 'INDEX'
					WHEN 'S' THEN 'SEQUENCE'
					WHEN 'm' THEN 'MATERIALIZED VIEW'
					WHEN 'c' THEN 'COMPOSITE TYPE'
					WHEN 't' THEN 'TOAST TABLE'
					WHEN 'f' THEN 'FOREIGN TABLE'
					ELSE C.relkind::TEXT
				END AS type,
				R.rolname AS owner,
				C.relacl AS acl
			FROM
				pg_class C
			JOIN
				pg_namespace N ON N.oid = C.relnamespace
			JOIN
				pg_roles R ON R.oid = C.relowner
			WHERE
				N.nspname NOT LIKE 'pg_%' AND N.nspname != 'information_schema'
		) SELECT * FROM objects
	`
	objectGet  = objectSelect + `WHERE name = ${'name'} AND schema = ${'schema'}`
	objectList = `WITH q AS (` + objectSelect + `) SELECT * FROM q ${where} ${orderby}`
)
