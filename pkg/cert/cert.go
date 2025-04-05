package cert

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"

	// Packages
	schema "github.com/mutablelogic/go-server/pkg/cert/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Certificate
type Cert struct {
	Name    string  `json:"name"`              // Common Name
	Subject *uint64 `json:"subject,omitempty"` // Subject
	Signer  *Cert   `json:"signer,omitempty"`  // Signer

	// database-id for the certificate
	id *uint64

	// The private key and certificate
	priv any
	x509 x509.Certificate
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// Supported key types
	keyTypeRSA   = "RSA"
	keyTypeECDSA = "ECDSA"

	// DefaultBits is the default number of bits for a RSA private key
	defaultBits = 2048
)

const (
	PemTypePrivateKey  = "PRIVATE KEY"
	PemTypeCertificate = "CERTIFICATE"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new certificate
func New(opts ...Opt) (*Cert, error) {
	cert, err := apply(opts...)
	if err != nil {
		return nil, err
	}

	// Check for key
	if cert.priv == nil || cert.PublicKey() == nil {
		return nil, fmt.Errorf("missing private or public key")
	}

	// Set the NotBefore based on signer, if not set
	if cert.Signer != nil && cert.x509.NotBefore.IsZero() {
		if !cert.Signer.x509.NotBefore.IsZero() && cert.Signer.x509.NotBefore.After(cert.x509.NotBefore) {
			cert.x509.NotBefore = cert.Signer.x509.NotBefore
		}
	}

	// Check for expiry
	if cert.x509.NotAfter.IsZero() {
		return nil, fmt.Errorf("missing expiry date")
	}

	// Set random serial number if not set
	if cert.x509.SerialNumber == nil {
		if err := WithRandomSerial()(cert); err != nil {
			return nil, err
		}
	}

	// commonName is required, set the name from the common name
	if cert.x509.Subject.CommonName == "" {
		return nil, fmt.Errorf("missing commonName")
	}
	if cert.Name == "" {
		cert.Name = cert.x509.Subject.CommonName
	}

	// Create the certificate
	signer := cert.Signer
	if signer == nil {
		signer = cert
	} else {
		cert.x509.Issuer = signer.x509.Subject
	}
	if data, err := x509.CreateCertificate(rand.Reader, &cert.x509, &signer.x509, cert.PublicKey(), signer.priv); err != nil {
		return nil, err
	} else {
		cert.x509.Raw = data
	}

	// Return the certificate
	return cert, nil
}

// Read a certificate
func Read(r io.Reader) (*Cert, error) {
	cert := new(Cert)
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Read until EOF
	for len(data) > 0 {
		// Decode the PEM block
		block, rest := pem.Decode(data)
		if block == nil {
			return nil, fmt.Errorf("invalid PEM block")
		}

		// Parse the block
		switch block.Type {
		case PemTypeCertificate:
			c, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, err
			} else {
				cert.x509 = *c
			}
		case PemTypePrivateKey:
			key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			} else {
				cert.priv = key
			}
		default:
			return nil, fmt.Errorf("invalid PEM block type: %q", block.Type)
		}

		// Move to next block
		data = rest
	}

	// Set name from serial number if not set
	if cert.Name == "" {
		cert.Name = fmt.Sprintf("%x", cert.x509.SerialNumber)
	}

	// Return success
	return cert, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c Cert) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.CertMeta())
}

func (c Cert) String() string {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return metadata from a cert
func (c Cert) SubjectMeta() schema.NameMeta {
	fieldPtr := func(field []string) *string {
		if len(field) > 0 {
			return types.StringPtr(field[0])
		} else {
			return nil
		}
	}
	return schema.NameMeta{
		CommonName:    c.x509.Subject.CommonName,
		Org:           fieldPtr(c.x509.Subject.Organization),
		Unit:          fieldPtr(c.x509.Subject.OrganizationalUnit),
		Country:       fieldPtr(c.x509.Subject.Country),
		City:          fieldPtr(c.x509.Subject.Locality),
		State:         fieldPtr(c.x509.Subject.Province),
		StreetAddress: fieldPtr(c.x509.Subject.StreetAddress),
		PostalCode:    fieldPtr(c.x509.Subject.PostalCode),
	}
}

// Return metadata from a cert
func (c Cert) CertMeta() schema.CertMeta {
	keybytes := func(priv any) []byte {
		if key, err := x509.MarshalPKCS8PrivateKey(priv); err != nil {
			return nil
		} else {
			return key
		}
	}
	signer := func(signer *Cert) *string {
		if signer == nil {
			return nil
		} else {
			return types.StringPtr(signer.Name)
		}
	}
	return schema.CertMeta{
		Signer:    signer(c.Signer),
		Subject:   c.Subject,
		IsCA:      c.IsCA(),
		NotBefore: c.x509.NotBefore,
		NotAfter:  c.x509.NotAfter,
		Cert:      c.x509.Raw,
		Key:       keybytes(c.priv),
	}
}

// Return true if the certificate is a certificate authority
func (c *Cert) IsCA() bool {
	return c.x509.IsCA
}

// Return the private key, or nil
func (c *Cert) PrivateKey() any {
	return c.priv
}

// Return the public key, or nil
func (c *Cert) PublicKey() any {
	switch k := c.priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey)
	default:
		return nil
	}
}

// Output certificate as PEM format
func (c *Cert) Write(w io.Writer) error {
	return pem.Encode(w, &pem.Block{Type: PemTypeCertificate, Bytes: c.x509.Raw})
}

// Write the private key as PEM format
func (c *Cert) WritePrivateKey(w io.Writer) error {
	data, err := x509.MarshalPKCS8PrivateKey(c.priv)
	if err != nil {
		return err
	}
	return pem.Encode(w, &pem.Block{Type: PemTypePrivateKey, Bytes: data})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (c *Cert) keyType() string {
	switch c.priv.(type) {
	case *rsa.PrivateKey:
		return keyTypeRSA
	case *ecdsa.PrivateKey:
		return keyTypeECDSA
	default:
		return ""
	}
}

func (c *Cert) keySubtype() string {
	switch k := c.priv.(type) {
	case *rsa.PrivateKey:
		return fmt.Sprintf("%d", k.N.BitLen())
	case *ecdsa.PrivateKey:
		return k.Params().Name
	default:
		return ""
	}
}
