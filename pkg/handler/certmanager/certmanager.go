package certmanager

import ( // Packages
	// Namespace imports
	"crypto/x509/pkix"
	"errors"
	"os"
	"time"

	. "github.com/djthorpe/go-errors"
	"github.com/mutablelogic/go-server/pkg/handler/certmanager/cert"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type certmanager struct {
	name  X509Name
	store CertStorage
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new auth task from the configuration
func New(c Config) (*certmanager, error) {
	task := new(certmanager)
	task.name = c.X509Name

	// Set storage for certificates
	if c.CertStorage == nil {
		return nil, ErrInternalAppError.With("missing 'CertStorage'")
	} else {
		task.store = c.CertStorage
	}

	// Return success
	return task, nil
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// List all certificates
func (task *certmanager) List() []Cert {
	certs, err := task.store.List()
	if err != nil {
		return nil
	}
	return certs
}

// Return a certificate by serial number
func (task *certmanager) Read(serial string) (Cert, error) {
	if cert, err := task.store.Read(serial); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotFound.With(serial)
		} else {
			return nil, err
		}
	} else {
		return cert, nil
	}
}

// Delete a certificate
func (task *certmanager) Delete(cert Cert) error {
	return task.store.Delete(cert)
}

// Create a new Certificate Authority
func (task *certmanager) CreateCA(commonName string, opts ...cert.Opt) (Cert, error) {
	// Default options
	o := []cert.Opt{
		cert.OptX509Name(pkix.Name{
			OrganizationalUnit: []string{task.name.OrganizationalUnit},
			Organization:       []string{task.name.Organization},
			Locality:           []string{task.name.Locality},
			Province:           []string{task.name.Province},
			Country:            []string{task.name.Country},
			StreetAddress:      []string{task.name.StreetAddress},
			PostalCode:         []string{task.name.PostalCode},
		}),
	}

	// Create the certificate and store it
	cert, err := cert.NewCA(commonName, append(o, opts...)...)
	if err != nil {
		return nil, err
	} else if err := task.store.Write(cert); err != nil {
		return nil, err
	}

	// Return success
	return cert, nil
}

// Create a new signed certificate. If ca is nil, the certificate is self-signed
func (task *certmanager) CreateSignedCert(commonName string, ca Cert, opts ...cert.Opt) (Cert, error) {
	// Default options
	o := []cert.Opt{
		cert.OptX509Name(pkix.Name{
			OrganizationalUnit: []string{task.name.OrganizationalUnit},
			Organization:       []string{task.name.Organization},
			Locality:           []string{task.name.Locality},
			Province:           []string{task.name.Province},
			Country:            []string{task.name.Country},
			StreetAddress:      []string{task.name.StreetAddress},
			PostalCode:         []string{task.name.PostalCode},
		}),
	}

	// We should make the ca "concrete" by reading it
	if ca != nil {
		var err error
		ca, err = task.Read(ca.Serial())
		if err != nil {
			return nil, err
		}
	}

	// Check for valid CA
	if ca != nil {
		if !ca.IsCA() {
			return nil, ErrBadParameter.With("Cannot sign without a valid CA")
		}
		if ca.Expires().Before(time.Now()) {
			return nil, ErrBadParameter.With("CA has expired")
		}
	}

	// Append KeyType to options
	if ca != nil {
		o = append(o, cert.OptKeyType(ca.KeyType()))
	}

	// Create the certificate and store it
	ca_, _ := ca.(*cert.Cert)
	cert, err := cert.NewCert(commonName, ca_, append(o, opts...)...)
	if err != nil {
		return nil, err
	} else if err := task.store.Write(cert); err != nil {
		return nil, err
	}

	// Return success
	return cert, nil
}
