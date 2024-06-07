package certmanager

import (
	"io"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	cert "github.com/mutablelogic/go-server/pkg/handler/certmanager/cert"
)

// Cert interface represents a certificate or certificate authority
type Cert interface {
	// Return Serial of the certificate
	Serial() string

	// Return the subject of the certificate
	Subject() string

	// Return ErrExpired if the certificate has expired,
	// or nil if the certificate is valid. Other error returns
	// indicate other problems with the certificate
	IsValid() error

	// Return the expiry date of the certificate
	Expires() time.Time

	// Return true if the certificate is a CA
	IsCA() bool

	// Return the key type
	KeyType() string

	// Write a .pem file with the certificate
	WriteCertificate(w io.Writer) error

	// Write a .pem file with the private key
	WritePrivateKey(w io.Writer) error
}

// CertStorage interface represents a storage for certificates
type CertStorage interface {
	server.Task

	// Return all certificates. This may not return the certificates
	// themselves, but the metadata for the certificates. Use Read
	// to get the certificate itself
	List() ([]Cert, error)

	// Read a certificate by serial number
	Read(string) (Cert, error)

	// Write a certificate
	Write(Cert) error

	// Delete a certificate
	Delete(Cert) error
}

// Ensure that Cert implements the certmanager.Cert interface
var _ Cert = (*cert.Cert)(nil)
