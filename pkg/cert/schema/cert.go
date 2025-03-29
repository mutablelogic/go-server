package schema

import (
	"context"
	"encoding/json"
	"net"
	"time"

	pg "github.com/djthorpe/go-pg"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Certificate Metadata
type CertMeta struct {
	Name         string    `json:"name"`
	Signer       *string   `json:"signer,omitempty"`
	SerialNumber string    `json:"serial_number,omitempty"`
	Subject      string    `json:"subject,omitempty"`
	NotBefore    time.Time `json:"not_before,omitempty"`
	NotAfter     time.Time `json:"not_after,omitempty"`
	IPs          []net.IP  `json:"ips,omitempty"`
	Hosts        []string  `json:"hosts,omitempty"`
	IsCA         bool      `json:"is_ca,omitempty"`
	KeyType      string    `json:"key_type,omitempty"`
	KeyBits      string    `json:"key_subtype,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c CertMeta) String() string {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
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
			-- certificate
			"cert" BYTEA NOT NULL,
			-- private key
			"key" BYTEA NOT NULL,
			-- expiry
			"not_before" TIMESTAMP NOT NULL,
			"not_after" TIMESTAMP NOT NULL,
			-- ca
			"is_ca" BOOLEAN NOT NULL,
			-- timestamp
			"ts" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`
	certInsert = `
		INSERT INTO ${"schema"}.cert (
			name, subject, signer, cert, key, not_before, not_after, is_ca
		) VALUES (
		 	@name, @subject, @signer, @cert, @key, @not_before, @not_after, @is_ca
		) RETURNING
			name, subject, signer, cert, key, ts
	`
)
