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
	"time"

	// Packages
	schema "github.com/mutablelogic/go-server/pkg/cert/schema"
	"github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Certificate
type Cert struct {
	Name    string     `json:"name"`                // Common Name
	Subject *uint64    `json:"subject,omitempty"`   // Subject
	Signer  *Cert      `json:"signer,omitempty"`    // Signer
	Ts      *time.Time `json:"timestamp,omitempty"` // Timestamp

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

	// Set the NotBefore and NotAfter dates based on signer
	if cert.Signer != nil {
		if !cert.Signer.x509.NotBefore.IsZero() && cert.Signer.x509.NotBefore.After(cert.x509.NotBefore) {
			cert.x509.NotBefore = cert.Signer.x509.NotBefore
		}
		if !cert.Signer.x509.NotAfter.IsZero() && cert.Signer.x509.NotAfter.Before(cert.x509.NotAfter) {
			cert.x509.NotAfter = cert.Signer.x509.NotAfter
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

	// Set the name from the common name
	cert.Name = cert.x509.Subject.CommonName

	// Create the certificate
	signer := cert.Signer
	if signer == nil {
		signer = cert
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

	// Return success
	return cert, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c Cert) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Meta())
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
func (c Cert) Meta() schema.CertMeta {
	signerNamePtr := func() *string {
		if c.Signer == nil {
			return nil
		}
		return types.StringPtr(c.Signer.Name)
	}
	return schema.CertMeta{
		Name:         c.Name,
		Signer:       signerNamePtr(),
		SerialNumber: fmt.Sprintf("%x", c.x509.SerialNumber),
		Subject:      c.x509.Subject.String(),
		NotBefore:    c.x509.NotBefore,
		NotAfter:     c.x509.NotAfter,
		IPs:          c.x509.IPAddresses,
		Hosts:        c.x509.DNSNames,
		IsCA:         c.IsCA(),
		KeyType:      c.keyType(),
		KeyBits:      c.keySubtype(),
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
