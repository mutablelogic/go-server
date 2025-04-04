package schema

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	// Packages
	pg "github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type NameId uint64

type NameMeta struct {
	CommonName    string  `json:"commonName,omitempty"`
	Org           *string `json:"organizationName,omitempty"`
	Unit          *string `json:"organizationalUnit,omitempty"`
	Country       *string `json:"countryName,omitempty"`
	City          *string `json:"localityName,omitempty"`
	State         *string `json:"stateOrProvinceName,omitempty"`
	StreetAddress *string `json:"streetAddress,omitempty"`
	PostalCode    *string `json:"postalCode,omitempty"`
}

type Name struct {
	Id uint64 `json:"id"`
	NameMeta
	Ts      time.Time `json:"timestamp,omitzero"`
	Subject *string   `json:"subject,omitempty"`
}

type NameList struct {
	Count uint64 `json:"count"`
	Body  []Name `json:"body,omitempty"`
}

type NameListRequest struct {
	pg.OffsetLimit
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (n NameMeta) String() string {
	data, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (n Name) String() string {
	data, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (n NameList) String() string {
	data, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// SELECT

func (n NameId) Select(bind *pg.Bind, op pg.Op) (string, error) {
	if n == 0 {
		return "", httpresponse.ErrBadRequest.With("id is missing")
	} else {
		bind.Set("id", n)
	}

	// Return query
	switch op {
	case pg.Get:
		return nameGet, nil
	case pg.Update:
		return namePatch, nil
	case pg.Delete:
		return nameDelete, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("unsupported NameId operation %q", op)
	}
}

func (n NameListRequest) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Set empty where
	bind.Set("where", "")

	// Bind offset and limit
	n.OffsetLimit.Bind(bind, NameListLimit)

	// Return query
	switch op {
	case pg.List:
		return nameList, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("unsupported NameListRequest operation %q", op)
	}
}

////////////////////////////////////////////////////////////////////////////////
// WRITER

func (n NameMeta) Insert(bind *pg.Bind) (string, error) {
	if commonName := strings.TrimSpace(n.CommonName); commonName == "" {
		return "", httpresponse.ErrBadRequest.With("commonName is missing")
	} else {
		bind.Set("commonName", commonName)
	}
	// TODO

	// Return insert or replace
	return nameReplace, nil
}

func (n NameMeta) Update(bind *pg.Bind) error {
	bind.Del("patch")
	if name := strings.TrimSpace(n.CommonName); name != "" {
		bind.Append("patch", `"commonName" = `+bind.Set("commonName", name))
	}
	if n.Org != nil {
		bind.Append("patch", `"organizationName" = `+bind.Set("organizationName", types.TrimStringPtr(*n.Org)))
	}
	if n.Unit != nil {
		bind.Append("patch", `"organizationalUnit" = `+bind.Set("organizationalUnit", types.TrimStringPtr(*n.Unit)))
	}
	if n.Country != nil {
		bind.Append("patch", `"countryName" = `+bind.Set("countryName", types.TrimStringPtr(*n.Country)))
	}
	if n.City != nil {
		bind.Append("patch", `"localityName" = `+bind.Set("localityName", types.TrimStringPtr(*n.City)))
	}
	if n.State != nil {
		bind.Append("patch", `"stateOrProvinceName" = `+bind.Set("stateOrProvinceName", types.TrimStringPtr(*n.State)))
	}
	if n.StreetAddress != nil {
		bind.Append("patch", `"streetAddress" = `+bind.Set("streetAddress", types.TrimStringPtr(*n.StreetAddress)))
	}
	if n.PostalCode != nil {
		bind.Append("patch", `"postalCode" = `+bind.Set("postalCode", types.TrimStringPtr(*n.PostalCode)))
	}

	// Join the patch fields
	if patch := bind.Join("patch", ", "); patch == "" {
		return httpresponse.ErrBadRequest.With("nothing to update")
	} else {
		bind.Set("patch", patch)
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// READER

func (n *Name) Scan(row pg.Row) error {
	return row.Scan(&n.Id, &n.CommonName, &n.Org, &n.Unit, &n.Country, &n.City, &n.State, &n.StreetAddress, &n.PostalCode, &n.Ts)
}

func (n *NameList) Scan(row pg.Row) error {
	var name Name
	if err := name.Scan(row); err != nil {
		return err
	} else {
		n.Body = append(n.Body, name)
	}
	return nil
}

func (n *NameList) ScanCount(row pg.Row) error {
	return row.Scan(&n.Count)
}

////////////////////////////////////////////////////////////////////////////////
// SQL

// Create name objects in the schema
func bootstrapName(ctx context.Context, conn pg.Conn) error {
	q := []string{
		nameCreateTable,
	}
	for _, query := range q {
		if err := conn.Exec(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

const (
	nameCreateTable = `
		CREATE TABLE IF NOT EXISTS ${"schema"}."name" (
			"id" SERIAL PRIMARY KEY,
			"commonName" TEXT NOT NULL,
			"organizationName" TEXT,
			"organizationalUnit" TEXT,
			"countryName" TEXT,
			"localityName" TEXT,			
			"stateOrProvinceName" TEXT,
			"streetAddress" TEXT,
			"postalCode" TEXT,
			"ts" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE ("commonName")
		)
	`
	nameReplace = `
		INSERT INTO ${"schema"}."name" (
			"commonName",  "organizationName", "organizationalUnit", "countryName", "localityName", "stateOrProvinceName", "streetAddress", "postalCode"
		) VALUES (
		 	@commonName, @organizationName, @organizationalUnit, @countryName, @localityName, @stateOrProvinceName, @streetAddress, @postalCode
		) ON CONFLICT ("commonName") DO UPDATE SET
			"organizationName" = @organizationName,
			"organizationalUnit" = @organizationalUnit,
			"countryName" = @countryName,
			"localityName" = @localityName,
			"stateOrProvinceName" = @stateOrProvinceName,
			"streetAddress" = @streetAddress,
			"postalCode" = @postalCode,
			"ts" = CURRENT_TIMESTAMP
		RETURNING
			"id", "commonName", "organizationName", "organizationalUnit", "countryName", "localityName", "stateOrProvinceName", "streetAddress", "postalCode", "ts"
	`
	namePatch = `
		UPDATE ${"schema"}."name" SET
			${patch}, "ts" = CURRENT_TIMESTAMP
		WHERE 
			"id" = @id
		RETURNING
			"id", "commonName", "organizationName", "organizationalUnit", "countryName", "localityName", "stateOrProvinceName", "streetAddress", "postalCode", "ts"
	`
	nameDelete = `
		DELETE FROM ${"schema"}."name" WHERE 
			"id" = @id
		RETURNING
			"id", "commonName", "organizationName", "organizationalUnit", "countryName", "localityName", "stateOrProvinceName", "streetAddress", "postalCode", "ts"
	`
	nameSelect = `
		SELECT
			"id", "commonName", "organizationName", "organizationalUnit", "countryName", "localityName", "stateOrProvinceName", "streetAddress", "postalCode", "ts"
		FROM
			${"schema"}."name"
	`
	nameGet  = nameSelect + ` WHERE "id" = @id`
	nameList = `WITH q AS (` + nameSelect + `) SELECT * FROM q ${where}`
)
