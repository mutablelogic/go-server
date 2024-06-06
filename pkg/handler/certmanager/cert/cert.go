package cert

// Ref: https://go.dev/src/crypto/tls/generate_cert.go

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"time"

	// Packages
	"github.com/mutablelogic/go-server/pkg/handler/certmanager"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Cert struct {
	data       []byte
	privateKey any
}

type keyType int

const (
	_ keyType = iota
	ED25519
	RSA2048
	P224
	P256
	P384
	P521
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	serialNumberLimit = new(big.Int).Lsh(big.NewInt(1), 128)
)

const (
	defaultYearsCA    = 2
	defaultMonthsCert = 3
	defaultKey        = RSA2048
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new certificate authority with the given configuration
// and options
func NewCA(c certmanager.Config, opt ...Opt) (*Cert, error) {
	var o opts

	// Set defaults
	o.KeyType = defaultKey
	o.Years = defaultYearsCA

	// Set options
	for _, fn := range opt {
		if err := fn(&o); err != nil {
			return nil, err
		}
	}

	// Get serial number
	var serial *big.Int
	if o.Serial != 0 {
		serial = big.NewInt(o.Serial)
	} else if serial = SerialNumber(); serial == nil {
		return nil, ErrInternalAppError.With("SerialNumber")
	}

	// Create a new certificate with a template
	template := x509TemplateFor(c, o, serial)
	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageCertSign

	// Generate public, private keys
	publicKey, privateKey, err := generateKey(o.KeyType)
	if err != nil {
		return nil, err
	}

	// Only RSA subject keys should have the KeyEncipherment KeyUsage bits set. In
	// the context of TLS this KeyUsage is particular to RSA key exchange and
	// authentication.
	if _, isRSA := privateKey.(*rsa.PrivateKey); isRSA {
		template.KeyUsage |= x509.KeyUsageKeyEncipherment
	}

	// Create self-signed CA
	cert := new(Cert)
	data, err := x509.CreateCertificate(rand.Reader, template, template, publicKey, privateKey)
	if err != nil {
		return nil, err
	} else {
		cert.data = data
		cert.privateKey = privateKey
	}

	// Return success
	return cert, nil
}

// Create a new certificate, either self-signed (if ca is nil) or
// signed by the certificate authority with the given options
// TODO: Add self-signing ability
func NewCert(ca *Cert, opt ...Opt) (*Cert, error) {
	var o opts

	// Set defaults
	o.KeyType = defaultKey
	o.Months = defaultMonthsCert

	// Set options
	for _, fn := range opt {
		if err := fn(&o); err != nil {
			return nil, err
		}
	}

	// Get serial number
	var serial *big.Int
	if o.Serial != 0 {
		serial = big.NewInt(o.Serial)
	} else if serial = SerialNumber(); serial == nil {
		return nil, ErrInternalAppError.With("SerialNumber")
	}

	parent, err := x509.ParseCertificate(ca.data)
	if err != nil {
		return nil, err
	}
	template, err := x509.ParseCertificate(ca.data)
	if err != nil {
		return nil, err
	}
	if o.Name != nil {
		template.Subject = *o.Name
	}
	if len(o.IPAddresses) > 0 {
		template.IPAddresses = o.IPAddresses
	}
	if len(o.DNSNames) > 0 {
		template.DNSNames = o.DNSNames
	}
	template.SerialNumber = serial
	template.NotBefore = time.Now()
	template.NotAfter = time.Now().AddDate(o.Years, o.Months, o.Days)
	template.IsCA = false
	template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
	// ECDSA, ED25519 and RSA subject keys should have the DigitalSignature
	// KeyUsage bits set in the x509.Certificate template
	template.KeyUsage = x509.KeyUsageDigitalSignature

	// Only RSA subject keys should have the KeyEncipherment KeyUsage bits set. In
	// the context of TLS this KeyUsage is particular to RSA key exchange and
	// authentication.
	if _, isRSA := ca.privateKey.(*rsa.PrivateKey); isRSA {
		template.KeyUsage |= x509.KeyUsageKeyEncipherment
	}

	// Generate public, private keys
	publicKey, privateKey, err := generateKey(o.KeyType)
	if err != nil {
		return nil, err
	}

	// Create cert signed by the CA
	cert := new(Cert)
	data, err := x509.CreateCertificate(rand.Reader, template, parent, publicKey, ca.privateKey)
	if err != nil {
		return nil, err
	} else {
		cert.privateKey = privateKey
		cert.data = data
	}

	// Return success
	return cert, nil
}

// Import certificate from byte stream
func NewFromBytes(data []byte) (*Cert, error) {
	public, rest := pem.Decode(data)
	if public == nil {
		return nil, ErrBadParameter.With("unable to decode certificate")
	}
	priv, _ := pem.Decode(rest)
	if priv == nil {
		return nil, ErrBadParameter.With("unable to decode private key")
	}

	cert := new(Cert)
	cert.data = public.Bytes
	if privKey, err := x509.ParsePKCS8PrivateKey(priv.Bytes); err != nil {
		return nil, err
	} else {
		cert.privateKey = privKey
	}

	// Return success
	return cert, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c *Cert) String() string {
	cert, err := x509.ParseCertificate(c.data)
	if err != nil {
		return fmt.Sprintf("{ %q: %q }", "error", err.Error())
	}
	v := struct {
		KeyType     string    `json:"key_type"`
		Serial      string    `json:"serial"`
		Subject     string    `json:"subject"`
		IsCA        bool      `json:"is_ca,omitempty"`
		NotBefore   time.Time `json:"not_before"`
		NotAfter    time.Time `json:"not_after"`
		IPAddresses []net.IP  `json:"ip_addresses,omitempty"`
		DNSNames    []string  `json:"dns_names,omitempty"`
	}{
		KeyType:     c.KeyType(),
		Serial:      cert.SerialNumber.String(),
		Subject:     cert.Subject.String(),
		IsCA:        cert.IsCA,
		NotBefore:   cert.NotBefore,
		NotAfter:    cert.NotAfter,
		IPAddresses: cert.IPAddresses,
		DNSNames:    cert.DNSNames,
	}
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Create a new serial number for the certificate, or nil if there
// was an error
func SerialNumber() *big.Int {
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil
	} else {
		return serialNumber
	}
}

// Return the key type
func (c *Cert) KeyType() string {
	switch v := c.privateKey.(type) {
	case *rsa.PrivateKey:
		return fmt.Sprintf("RSA%d", v.Size()*8)
	case *ecdsa.PrivateKey:
		return "ECDSA " + v.Curve.Params().Name
	case ed25519.PrivateKey:
		return "ED25519"
	default:
		return "UNKNOWN"
	}
}

// Write a .pem file with the certificate
func (c *Cert) WriteCertificate(w io.Writer) error {
	return pem.Encode(w, &pem.Block{Type: "CERTIFICATE", Bytes: c.data})
}

// Write a .pem file with the private key
func (c *Cert) WritePrivateKey(w io.Writer) error {
	if privBytes, err := x509.MarshalPKCS8PrivateKey(c.privateKey); err != nil {
		return err
	} else if err := pem.Encode(w, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return err
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func x509TemplateFor(c certmanager.Config, o opts, serial *big.Int) *x509.Certificate {
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization:       []string{c.X509Name.Organization},
			OrganizationalUnit: []string{c.X509Name.OrganizationalUnit},
			Country:            []string{c.X509Name.Country},
			Locality:           []string{c.X509Name.Locality},
			Province:           []string{c.X509Name.Province},
			StreetAddress:      []string{c.X509Name.StreetAddress},
			PostalCode:         []string{c.X509Name.PostalCode},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(o.Years, o.Months, o.Days),
		ExtKeyUsage:           []x509.ExtKeyUsage{},
		BasicConstraintsValid: true,
	}

	// Set X509 Name
	if o.Name != nil {
		template.Subject = *o.Name
	}

	// Return the template
	return template
}

// ECDSA curve to use to generate a key. Valid values are P224, P256 (default), P384, P521
// If empty, RSA keys will be generated instead
func generateKey(t keyType) (any, any, error) {
	switch t {
	case ED25519:
		return ed25519.GenerateKey(rand.Reader)
	case RSA2048:
		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		return &priv.PublicKey, priv, err
	case P224:
		priv, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
		return &priv.PublicKey, priv, err
	case P256:
		priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		return &priv.PublicKey, priv, err
	case P384:
		priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		return &priv.PublicKey, priv, err
	case P521:
		priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
		return &priv.PublicKey, priv, err
	default:
		return nil, nil, ErrBadParameter
	}
}
