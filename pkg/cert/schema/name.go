package schema

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	// Packages
	pg "github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-service/pkg/httpresponse"
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
	default:
		return "", httpresponse.ErrInternalError.Withf("unsupported NameId operation %q", op)
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

	// Return insert or replace
	return nameReplace, nil
}

func (n NameMeta) Update(bind *pg.Bind) error {
	return httpresponse.ErrNotImplemented.With("NameMeta.Update")
}

////////////////////////////////////////////////////////////////////////////////
// READER

func (n *Name) Scan(row pg.Row) error {
	return row.Scan(&n.Id, &n.CommonName, &n.Org, &n.Unit, &n.Country, &n.City, &n.State, &n.StreetAddress, &n.PostalCode, &n.Ts)
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
	nameList = `WITH q AS (` + nameSelect + `) SELECT * FROM q ${where} ${offsetlimit}`
)
