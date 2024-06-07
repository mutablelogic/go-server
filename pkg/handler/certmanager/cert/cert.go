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

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Cert represents a certificate with a private key which can be used for
// signing other certificates
type Cert struct {
	data       []byte
	privateKey any
}

type keyType int

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	_ keyType = iota
	ED25519
	RSA2048
	P224
	P256
	P384
	P521
)

const (
	defaultYearsCA    = 2
	defaultMonthsCert = 3
	defaultKey        = RSA2048
)

var (
	// Maximum is 128 bits
	serialNumberLimit = new(big.Int).Lsh(big.NewInt(1), 128)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new certificate authority with the given options
func NewCA(commonName string, opt ...Opt) (*Cert, error) {
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
	template := &x509.Certificate{
		SerialNumber:          serial,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(o.Years, o.Months, o.Days),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{},
		BasicConstraintsValid: true,
	}

	// Set subject
	if o.Name != nil {
		template.Subject = *o.Name
	} else {
		template.Subject = pkix.Name{}
	}
	template.Subject.CommonName = commonName

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
func NewCert(commonName string, ca *Cert, opt ...Opt) (*Cert, error) {
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

	// Parse the CA certificate
	var parent *x509.Certificate
	if ca != nil {
		var err error
		parent, err = x509.ParseCertificate(ca.data)
		if err != nil {
			return nil, err
		}
		if !parent.IsCA {
			return nil, ErrBadParameter.With("Invalid CA certificate")
		}
		if parent.NotAfter.Before(time.Now()) {
			return nil, ErrBadParameter.With("CA certificate has expired")
		}
		if parent.NotBefore.After(time.Now()) {
			return nil, ErrBadParameter.With("CA certificate is not yet valid")
		}
	}

	template := &x509.Certificate{
		SerialNumber:          serial,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(o.Years, o.Months, o.Days),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	// Set subject
	if o.Name != nil {
		template.Subject = *o.Name
	} else if parent != nil {
		template.Subject = parent.Subject
	}

	// Set common name
	template.Subject.CommonName = commonName

	// Set IP Addresses and DNS Names
	if len(o.IPAddresses) > 0 {
		template.IPAddresses = o.IPAddresses
	}
	if len(o.DNSNames) > 0 {
		template.DNSNames = o.DNSNames
	}

	// Generate public, private keys
	publicKey, privateKey, err := generateKey(o.KeyType)
	if err != nil {
		return nil, err
	}

	// Who is going to sign the certificate?
	signer, signerPrivateKey := template, privateKey
	if parent != nil {
		signer, signerPrivateKey = parent, ca.privateKey
	}

	// Only RSA subject keys should have the KeyEncipherment KeyUsage bits set. In
	// the context of TLS this KeyUsage is particular to RSA key exchange and
	// authentication.
	if _, isRSA := signerPrivateKey.(*rsa.PrivateKey); isRSA {
		template.KeyUsage |= x509.KeyUsageKeyEncipherment
	}

	// Set authority key id
	if parent != nil {
		template.AuthorityKeyId = parent.SubjectKeyId
	}

	// Create cert signed by the CA or self
	cert := new(Cert)
	data, err := x509.CreateCertificate(rand.Reader, template, signer, publicKey, signerPrivateKey)
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

func (c *Cert) MarshalJSON() ([]byte, error) {
	type resp struct {
		Serial         string    `json:"serial"`
		KeyType        string    `json:"key_type"`
		CommonName     string    `json:"name"`
		IsCA           bool      `json:"is_ca,omitempty"`
		NotBefore      time.Time `json:"not_before"`
		NotAfter       time.Time `json:"not_after"`
		IPAddresses    []net.IP  `json:"ip_addresses,omitempty"`
		DNSNames       []string  `json:"dns_names,omitempty"`
		SubjectKeyId   []byte    `json:"subject_key_id,omitempty"`
		AuthorityKeyId []byte    `json:"authority_key_id,omitempty"`
	}
	type error struct {
		Error string `json:"error"`
	}

	cert, err := x509.ParseCertificate(c.data)
	if err != nil {
		return json.Marshal(error{Error: err.Error()})
	} else {
		return json.Marshal(resp{
			Serial:         cert.SerialNumber.String(),
			KeyType:        c.KeyType(),
			CommonName:     cert.Subject.CommonName,
			IsCA:           cert.IsCA,
			NotBefore:      cert.NotBefore,
			NotAfter:       cert.NotAfter,
			IPAddresses:    cert.IPAddresses,
			DNSNames:       cert.DNSNames,
			SubjectKeyId:   cert.SubjectKeyId,
			AuthorityKeyId: cert.AuthorityKeyId,
		})
	}
}

func (c *Cert) String() string {
	data, _ := json.MarshalIndent(c, "", "  ")
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

func (c *Cert) IsCA() bool {
	cert, err := x509.ParseCertificate(c.data)
	if err != nil {
		return false
	}
	return cert.IsCA
}

func (c *Cert) Serial() string {
	cert, err := x509.ParseCertificate(c.data)
	if err != nil {
		return ""
	}
	return cert.SerialNumber.String()
}

func (c *Cert) Subject() string {
	cert, err := x509.ParseCertificate(c.data)
	if err != nil {
		return ""
	}
	return cert.Subject.CommonName
}

func (c *Cert) Expires() time.Time {
	if cert, err := x509.ParseCertificate(c.data); err != nil {
		return time.Time{}
	} else {
		return cert.NotAfter
	}
}

func (c *Cert) IsValid() error {
	cert, err := x509.ParseCertificate(c.data)
	if err != nil {
		return err
	}
	if _, err := cert.Verify(x509.VerifyOptions{}); err != nil {
		return err
	}
	// Return success
	return nil
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
