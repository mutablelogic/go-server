package schema

import (
	"encoding/json"
	"path/filepath"
	"strings"

	// Packages
	pg "github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type TablespaceName string

type TablespaceMeta struct {
	Name  *string `json:"name,omitempty" help:"Tablespace name"`
	Owner *string `json:"owner,omitempty" help:"Owner"`
	Acl   ACLList `json:"acl,omitempty" help:"Access privileges"`
}

type Tablespace struct {
	Oid uint32 `json:"oid"`
	TablespaceMeta
	Location string   `json:"location,omitempty" help:"Location"`
	Options  []string `json:"options,omitempty" help:"Options"`
	Size     uint64   `json:"bytes,omitempty" help:"Size of schema in bytes"`
}

type TablespaceListRequest struct {
	pg.OffsetLimit
}

type TablespaceList struct {
	Count uint64       `json:"count"`
	Body  []Tablespace `json:"body,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t TablespaceMeta) String() string {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (t Tablespace) String() string {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (t TablespaceListRequest) String() string {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (t TablespaceList) String() string {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// SELECT

func (t TablespaceListRequest) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Order
	bind.Set("orderby", `ORDER BY name ASC`)

	// Bind offset and limit
	t.OffsetLimit.Bind(bind, TablespaceListLimit)

	// Where
	bind.Set("where", ``)

	// Return query
	switch op {
	case pg.List:
		return tablespaceList, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("unsupported TablespaceListRequest operation %q", op)
	}
}

func (t TablespaceName) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Set name
	if name := strings.TrimSpace(string(t)); name == "" {
		return "", httpresponse.ErrBadRequest.With("tablespace name is missing")
	} else {
		bind.Set("name", name)
	}

	// Return query
	switch op {
	case pg.Get:
		return tablespaceGet, nil
	case pg.Update:
		return tablespaceRename, nil
	case pg.Delete:
		return tablespaceDelete, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("unsupported TablespaceName operation %q", op)
	}
}

////////////////////////////////////////////////////////////////////////////////
// READER

func (t *Tablespace) Scan(row pg.Row) error {
	var priv []string
	t.Acl = ACLList{}
	if err := row.Scan(&t.Oid, &t.Name, &t.Owner, &priv, &t.Options, &t.Size, &t.Location); err != nil {
		return err
	}
	for _, v := range priv {
		item, err := NewACLItem(v)
		if err != nil {
			return err
		}
		t.Acl.Append(item)
	}
	return nil
}

func (t *TablespaceList) Scan(row pg.Row) error {
	var tablespace Tablespace
	if err := tablespace.Scan(row); err != nil {
		return err
	} else {
		t.Body = append(t.Body, tablespace)
	}
	return nil
}

func (t *TablespaceList) ScanCount(row pg.Row) error {
	return row.Scan(&t.Count)
}

////////////////////////////////////////////////////////////////////////////////
// WRITER

func (t TablespaceMeta) Insert(bind *pg.Bind) (string, error) {
	// Set name
	if name := strings.TrimSpace(types.PtrString(t.Name)); name == "" {
		return "", httpresponse.ErrBadRequest.With("name is missing")
	} else if strings.HasPrefix(name, reservedPrefix) {
		return "", httpresponse.ErrBadRequest.Withf("cannot create a tablespace prefixed with %q", reservedPrefix)
	} else {
		bind.Set("name", name)
	}

	// Set location
	if location, ok := bind.Get("location").(string); !ok {
		return "", httpresponse.ErrBadRequest.With("location is missing")
	} else if location = strings.TrimSpace(location); location == "" {
		return "", httpresponse.ErrBadRequest.With("location is missing")
	} else if !filepath.IsAbs(location) {
		return "", httpresponse.ErrBadRequest.Withf("location %q is not absolute", location)
	} else {
		bind.Set("location", location)
	}

	// Set with
	if with, err := t.with(); err != nil {
		return "", err
	} else {
		bind.Set("with", with)
	}

	// Return success
	return tablespaceCreate, nil
}

func (t TablespaceMeta) Update(bind *pg.Bind) error {
	return httpresponse.ErrNotImplemented.With("TablespaceMeta.Update")
}

func (t TablespaceName) Insert(bind *pg.Bind) (string, error) {
	return "", httpresponse.ErrNotImplemented.With("TablespaceName.Insert")
}

func (t TablespaceName) Update(bind *pg.Bind) error {
	if name := strings.TrimSpace(string(t)); name == "" {
		return httpresponse.ErrBadRequest.With("name is missing")
	} else if strings.HasPrefix(name, reservedPrefix) {
		return httpresponse.ErrBadRequest.Withf("cannot create a tablespace prefixed with %q", reservedPrefix)
	} else {
		bind.Set("old_name", name)
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (t TablespaceMeta) with() (string, error) {
	var with []string

	// Owner
	if owner := strings.TrimSpace(types.PtrString(t.Owner)); owner != "" {
		with = append(with, `OWNER `+types.DoubleQuote(owner))
	}

	if len(with) > 0 {
		return strings.Join(with, " "), nil
	} else {
		return "", nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// SQL

const (
	tablespaceSelect = `
		WITH t AS (
			SELECT
				T.oid AS "oid", T.spcname AS "name", R.rolname AS "owner", T.spcacl AS "acl",  T.spcoptions AS "options", pg_tablespace_size(T.oid) AS "size",
				CASE
					WHEN T.spcname = 'pg_global' THEN (SELECT setting||'/base' FROM pg_settings WHERE name='data_directory')
					WHEN T.spcname = 'pg_default' THEN (SELECT setting||'/global' FROM pg_settings WHERE name='data_directory')
					ELSE pg_tablespace_location(T.oid)
				END AS "location"
			FROM
				"pg_catalog"."pg_tablespace" T
			JOIN
				"pg_catalog"."pg_roles" R ON T.spcowner = R.oid			
		) SELECT * FROM t`
	tablespaceGet    = tablespaceSelect + ` WHERE "name" = ${'name'}`
	tablespaceList   = `WITH q AS (` + tablespaceSelect + `) SELECT * FROM q ${where} ${orderby}`
	tablespaceCreate = `CREATE TABLESPACE ${"name"} ${with} LOCATION ${'location'}`
	tablespaceRename = `ALTER TABLESPACE ${"old_name"} RENAME TO ${"name"}`
	tablespaceDelete = `DROP TABLESPACE ${"name"}`
)
