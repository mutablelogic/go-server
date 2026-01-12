package schema

import (
	"context"
	"encoding/json"
	"math/big"
	"strings"
	"time"

	// Packages
	pg "github.com/mutablelogic/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Certificate Name
type CertName string

// Certificate Metadata for creating a new certificate
type CertCreateMeta struct {
	Name         string        `json:"name,omitempty"`
	CommonName   string        `json:"common_name,omitempty"`
	Signer       string        `json:"signer,omitempty"`
	Subject      string        `json:"subject,omitempty"`
	SerialNumber *big.Int      `json:"serial_number,omitempty"`
	Expiry       time.Duration `json:"expiry,omitempty"`
	IsCA         bool          `json:"is_ca,omitempty"`
	KeyType      string        `json:"key_type,omitempty"`
	Address      []string      `json:"address,omitempty"`
}

// Certificate Metadata
type CertMeta struct {
	Signer    *string   `json:"signer,omitempty"`
	Subject   *uint64   `json:"subject,omitempty"`
	NotBefore time.Time `json:"not_before,omitzero"`
	NotAfter  time.Time `json:"not_after,omitzero"`
	IsCA      bool      `json:"is_ca,omitempty"`
	Cert      []byte    `json:"cert,omitempty"`
	Key       []byte    `json:"key,omitempty"`
}

// Certificate Metadata
type Cert struct {
	Name string `json:"name"`
	CertMeta
	Ts time.Time `json:"timestamp,omitzero"`
}

type CertList struct {
	Count uint64     `json:"count"`
	Body  []CertMeta `json:"body,omitempty"`
}

type CertListRequest struct {
	pg.OffsetLimit
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c Cert) String() string {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (c CertMeta) String() string {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (c CertListRequest) String() string {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (c CertList) String() string {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// DELECTOR

func (c CertName) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Name
	if name := string(c); name == "" {
		return "", httpresponse.ErrBadRequest.With("name is missing")
	} else if !types.IsIdentifier(name) {
		return "", httpresponse.ErrBadRequest.With("name is invalid")
	} else {
		bind.Set("name", name)
	}

	switch op {
	case pg.Get:
		return certGet, nil
	case pg.Delete:
		return certDelete, nil
	default:
		return "", httpresponse.ErrBadRequest.Withf("CertName: operation %q is not supported", op)
	}
}

////////////////////////////////////////////////////////////////////////////////
// READER

func (c *Cert) Scan(row pg.Row) error {
	// Scan the row
	if err := row.Scan(&c.Name, &c.Subject, &c.Signer, &c.Cert, &c.Key, &c.NotBefore, &c.NotAfter, &c.IsCA, &c.Ts); err != nil {
		return err
	}
	// Todo
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// WRITER

func (c CertMeta) Insert(bind *pg.Bind) (string, error) {
	// Name
	if !bind.Has("name") {
		return "", httpresponse.ErrBadRequest.With("name is missing")
	} else if name, ok := bind.Get("name").(string); !ok {
		return "", httpresponse.ErrBadRequest.With("name is invalid")
	} else if name = strings.TrimSpace(name); name == "" {
		return "", httpresponse.ErrBadRequest.With("name is missing")
	} else if !types.IsIdentifier(name) {
		return "", httpresponse.ErrBadRequest.With("name is invalid")
	} else {
		bind.Set("name", name)
	}

	// Signer
	bind.Set("signer", c.Signer)

	// Subject
	if subject := types.PtrUint64(c.Subject); subject == 0 {
		return "", httpresponse.ErrBadRequest.With("subject is missing")
	} else {
		bind.Set("subject", subject)
	}

	// NotBefore
	if c.NotBefore.IsZero() {
		return "", httpresponse.ErrBadRequest.With("not_before is missing")
	} else {
		bind.Set("not_before", c.NotBefore)
	}

	// NotAfter
	if c.NotAfter.IsZero() {
		return "", httpresponse.ErrBadRequest.With("not_after is missing")
	} else if c.NotAfter.Before(c.NotBefore) {
		return "", httpresponse.ErrBadRequest.With("not_after is before not_before")
	} else {
		bind.Set("not_after", c.NotAfter)
	}

	// Set cert and key
	bind.Set("cert", c.Cert)
	bind.Set("key", c.Key)

	// IsCA
	bind.Set("is_ca", c.IsCA)

	// Return insert or replace
	return certReplace, nil
}

func (c CertMeta) Update(bind *pg.Bind) error {
	return httpresponse.ErrNotImplemented
}

////////////////////////////////////////////////////////////////////////////////
// SQL

// Create objects in the schema
func bootstrapCert(ctx context.Context, conn pg.Conn) error {
	q := []string{
		certCreateTable,
	}
	for _, query := range q {
		if err := conn.Exec(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

const (
	certCreateTable = `
		CREATE TABLE IF NOT EXISTS ${"schema"}.cert (
			-- cert name
			"name" TEXT PRIMARY KEY,
			-- subject
			"subject" SERIAL REFERENCES ${"schema"}."name"("id") ON DELETE CASCADE,
			-- signer
			"signer" TEXT REFERENCES ${"schema"}.cert("name") ON DELETE RESTRICT,
			-- expiry
			"not_before" TIMESTAMP NOT NULL,
			"not_after" TIMESTAMP NOT NULL,
			-- ca
			"is_ca" BOOLEAN NOT NULL,
			-- certificate
			"cert" BYTEA NOT NULL,
			-- private key
			"key" BYTEA NOT NULL,
			-- timestamp
			"ts" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`
	certReplace = `
		INSERT INTO ${"schema"}.cert (
			name, subject, signer, cert, key, not_before, not_after, is_ca
		) VALUES (
		 	@name, @subject, @signer, @cert, @key, @not_before, @not_after, @is_ca
		) ON CONFLICT (name) DO UPDATE SET
			subject = @subject,
			signer = @signer,
			cert = @cert,
			key = @key,
			not_before = @not_before,
			not_after = @not_after,
			is_ca = @is_ca,
			ts = CURRENT_TIMESTAMP
		RETURNING
			name, subject, signer, cert, key, not_before, not_after, is_ca, ts
	`
	certDelete = `
		DELETE FROM ${"schema"}.cert WHERE 
			name = @name
		RETURNING
			name, subject, signer, cert, key, not_before, not_after, is_ca, ts
	`
	certSelect = `
		SELECT
			name, subject, signer, cert, key, not_before, not_after, is_ca, ts
		FROM ${"schema"}.cert
	`
	certGet = certSelect + `WHERE name = @name`
)
