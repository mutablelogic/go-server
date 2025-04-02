package cert

import (
	"context"

	// Packages
	pg "github.com/djthorpe/go-pg"
	schema "github.com/mutablelogic/go-server/pkg/cert/schema"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type CertManager struct {
	conn pg.PoolConn
	root *Cert
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new certificate manager, with a root certificate authority
func NewCertManager(ctx context.Context, conn pg.PoolConn, opt ...Opt) (*CertManager, error) {
	self := new(CertManager)
	self.conn = conn.With("schema", schema.SchemaName).(pg.PoolConn)

	// Create the root certificate from options
	if root, err := New(opt...); err != nil {
		return nil, err
	} else if !root.IsCA() {
		return nil, httpresponse.ErrInternalError.With("root certificate must be a certificate authority")
	} else {
		self.root = root
	}

	// If the schema does not exist, then bootstrap it
	if err := self.conn.Tx(ctx, func(conn pg.Conn) error {
		if exists, err := pg.SchemaExists(ctx, conn, schema.SchemaName); err != nil {
			return err
		} else if !exists {
			return schema.Bootstrap(ctx, conn)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Register the root certificate
	if err := self.conn.Tx(ctx, func(conn pg.Conn) error {
		// Register the name
		subject, err := self.RegisterName(ctx, self.root.SubjectMeta())
		if err != nil {
			return err
		} else {
			self.root.Subject = types.Uint64Ptr(subject.Id)
		}

		// Get a cert for the root, if no cert exists then register a new one
		certs, err := self.ListCerts(ctx, schema.CertListRequest{
			Subject: subject.Id,
		})

		if cert, err := self.RegisterCert(ctx, self.root); err != nil {
			return err
		} else {
			self.root = cert
		}

		// Return success
		return nil
	}); err != nil {
		return nil, err
	}

	// Return success
	return self, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the root certificate
func (certmanager *CertManager) Root() *Cert {
	return certmanager.root
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (certmanager *CertManager) RegisterName(ctx context.Context, meta schema.NameMeta) (*schema.Name, error) {
	var name schema.Name
	if err := certmanager.conn.Insert(ctx, &name, meta); err != nil {
		return nil, err
	} else {
		return &name, nil
	}
}

func (certmanager *CertManager) RegisterCert(ctx context.Context, meta *Cert) (*Cert, error) {
	// Return error
	return nil, httpresponse.ErrNotImplemented
}

func (certmanager *CertManager) GetName(ctx context.Context, id uint64) (*schema.Name, error) {
	var name schema.Name
	if err := certmanager.conn.Get(ctx, &name, schema.NameId(id)); err != nil {
		return nil, err
	} else {
		return &name, nil
	}
}

func (certmanager *CertManager) UpdateName(ctx context.Context, id uint64, meta schema.NameMeta) (*schema.Name, error) {
	// Don't allow to update the commonName of the root certificate
	if id == types.PtrUint64(certmanager.root.Subject) {
		if meta.CommonName != "" {
			return nil, httpresponse.ErrConflict.With("cannot update commonName of root certificate")
		}
	}

	// Allow the update
	var name schema.Name
	if err := certmanager.conn.Update(ctx, &name, schema.NameId(id), meta); err != nil {
		return nil, err
	} else {
		return &name, nil
	}
}

func (certmanager *CertManager) DeleteName(ctx context.Context, id uint64) (*schema.Name, error) {
	var name schema.Name
	// Don't allow the root certificate to be deleted
	if id == types.PtrUint64(certmanager.root.Subject) {
		return nil, httpresponse.ErrConflict.With("cannot delete root certificate")
	} else if err := certmanager.conn.Delete(ctx, &name, schema.NameId(id)); err != nil {
		return nil, err
	} else {
		return &name, nil
	}
}

func (certmanager *CertManager) ListNames(ctx context.Context) (*schema.NameList, error) {
	var list schema.NameList
	if err := certmanager.conn.List(ctx, &list, schema.NameListRequest{}); err != nil {
		return nil, err
	} else {
		return &list, nil
	}
}
