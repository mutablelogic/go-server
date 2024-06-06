package certmanager

import "io"

// Cert interface represents a certificate or certificate authority
type Cert interface {
	// Return Serial of the certificate
	Serial() string

	// Return the subject of the certificate
	Subject() string

	// Return true if the certificate is valid
	IsValid() bool

	// Return true if the certificate is a CA
	IsCA() bool

	// Return the key type
	KeyType() string

	// Write a .pem file with the certificate
	WriteCertificate(w io.Writer) error

	// Write a .pem file with the private key
	WritePrivateKey(w io.Writer) error
}
