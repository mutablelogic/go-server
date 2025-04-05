package cert

import (
	"context"
	"errors"
	"fmt"

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

		// Register the cert
		if cert, err := self.RegisterCert(ctx, self.root.Name, self.root.CertMeta()); err != nil {
			return err
		} else {
			fmt.Println(cert)
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

// Create a certificate with the given name, and a signer. The certificate is created in the database
func (certmanager *CertManager) CreateCert(ctx context.Context, name string, opt ...Opt) (*schema.Cert, error) {
	// Create the certificate from options
	cert, err := New(append([]Opt{
		withName(name),
	}, opt...)...)
	if err != nil {
		return nil, err
	}

	// Check the certificate doesn't already exist with that name, then insert it
	var result *schema.Cert
	if err := certmanager.conn.Tx(ctx, func(conn pg.Conn) error {
		// Check if a certificate with that name already exists
		if _, err := certmanager.GetCert(ctx, name); errors.Is(err, pg.ErrNotFound) {
			// OK
		} else if err != nil {
			return err
		} else {
			return httpresponse.ErrConflict.Withf("Certificate with name %q already exists", name)
		}

		// Set the subject of the certificate based on the SubjectMeta
		subject, err := certmanager.RegisterName(ctx, cert.SubjectMeta())
		if err != nil {
			return err
		} else {
			cert.Subject = types.Uint64Ptr(subject.Id)
		}

		// Register the certificate
		if cert, err := certmanager.RegisterCert(ctx, name, cert.CertMeta()); err != nil {
			return err
		} else {
			result = cert
		}
		// Return success
		return nil
	}); err != nil {
		return nil, err
	}

	// Return the certificate
	return result, nil
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

func (certmanager *CertManager) RegisterCert(ctx context.Context, name string, meta schema.CertMeta) (*schema.Cert, error) {
	var cert schema.Cert
	if err := certmanager.conn.With("name", name).Insert(ctx, &cert, meta); err != nil {
		return nil, err
	} else {
		return &cert, nil
	}
}

func (certmanager *CertManager) GetName(ctx context.Context, id uint64) (*schema.Name, error) {
	var name schema.Name
	if err := certmanager.conn.Get(ctx, &name, schema.NameId(id)); err != nil {
		return nil, err
	} else {
		return &name, nil
	}
}

func (certmanager *CertManager) GetCert(ctx context.Context, name string) (*schema.Cert, error) {
	var cert schema.Cert
	if err := certmanager.conn.Get(ctx, &cert, schema.CertName(name)); err != nil {
		return nil, err
	} else {
		return &cert, nil
	}
}

func (certmanager *CertManager) UpdateName(ctx context.Context, id uint64, meta schema.NameMeta) (*schema.Name, error) {
	var name schema.Name

	// Don't allow to update the commonName of the root certificate
	if id == types.PtrUint64(certmanager.root.Subject) {
		if meta.CommonName != "" {
			return nil, httpresponse.ErrConflict.With("cannot update commonName of root certificate")
		}
	}
	if err := certmanager.conn.Update(ctx, &name, schema.NameId(id), meta); err != nil {
		return nil, err
	} else {
		return &name, nil
	}
}

func (certmanager *CertManager) UpdateCert(ctx context.Context, name string, meta schema.CertMeta) (*schema.Cert, error) {
	var cert schema.Cert

	// Don't allow to update the root certificate
	if name == certmanager.root.Name {
		return nil, httpresponse.ErrConflict.With("cannot update root certificate")
	}
	if err := certmanager.conn.Update(ctx, &cert, schema.CertName(name), meta); err != nil {
		return nil, err
	} else {
		return &cert, nil
	}
}

func (certmanager *CertManager) DeleteName(ctx context.Context, id uint64) (*schema.Name, error) {
	var name schema.Name

	// Don't allow the root certificate to be deleted
	if id == types.PtrUint64(certmanager.root.Subject) {
		return nil, httpresponse.ErrConflict.With("cannot delete root name")
	} else if err := certmanager.conn.Delete(ctx, &name, schema.NameId(id)); err != nil {
		return nil, err
	} else {
		return &name, nil
	}
}

func (certmanager *CertManager) DeleteCert(ctx context.Context, name string) (*schema.Cert, error) {
	var cert schema.Cert

	// Don't allow the root certificate to be deleted
	if name == certmanager.root.Name {
		return nil, httpresponse.ErrConflict.With("cannot delete root certificate")
	}
	if err := certmanager.conn.Delete(ctx, &cert, schema.CertName(name)); err != nil {
		return nil, err
	} else {
		return &cert, nil
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
