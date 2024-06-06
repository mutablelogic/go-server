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
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"strings"
	"time"

	// Packages
	"github.com/mutablelogic/go-server/pkg/handler/certmanager"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Cert struct {
	ca         bool
	serial     *big.Int
	privateKey any
	data       []byte
}

type KeyType int

const (
	_ KeyType = iota
	ED25519
	RSA2048
	P224
	P256
	P384
	P521
)

///////////////////////////////////////////////////////////////////////////////
// LIFE CYCLE

func New(c certmanager.Config, k KeyType, years, months, days int, ca bool, host string, opts ...Opts) (*Cert, error) {
	cert := new(Cert)

	// Random serial number
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	} else {
		cert.serial = serialNumber
	}

	// Create the certificate template
	template := &x509.Certificate{
		SerialNumber: cert.serial,
		Subject: pkix.Name{
			Organization:       []string{c.Organization},
			OrganizationalUnit: []string{c.OrganizationalUnit},
			Country:            []string{c.Country},
			Locality:           []string{c.Locality},
			Province:           []string{c.Province},
			StreetAddress:      []string{c.StreetAddress},
			PostalCode:         []string{c.PostalCode},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(years, months, days),
		IsCA:        ca,
		ExtKeyUsage: []x509.ExtKeyUsage{},
		// ECDSA, ED25519 and RSA subject keys should have the DigitalSignature
		// KeyUsage bits set in the x509.Certificate template
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	// Create new private key
	public, private, err := generateKey(k)
	if err != nil {
		return nil, err
	} else {
		cert.privateKey = private
	}

	// Only RSA subject keys should have the KeyEncipherment KeyUsage bits set. In
	// the context of TLS this KeyUsage is particular to RSA key exchange and
	// authentication.
	if _, isRSA := private.(*rsa.PrivateKey); isRSA {
		template.KeyUsage |= x509.KeyUsageKeyEncipherment
	}

	// Set hosts
	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	// Set CA
	if ca {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
		cert.ca = true
	} else {
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth)
	}

	data, err := x509.CreateCertificate(rand.Reader, template, template, public, private)
	if err != nil {
		return nil, err
	} else {
		cert.data = data
	}

	// Return success
	return cert, nil
}

/*
func LoadX509KeyPair(certFile, keyFile string) (*x509.Certificate, *rsa.PrivateKey) {
    cf, e := ioutil.ReadFile(certFile)
    if e != nil {
        fmt.Println("cfload:", e.Error())
        os.Exit(1)
    }

    kf, e := ioutil.ReadFile(keyFile)
    if e != nil {
        fmt.Println("kfload:", e.Error())
        os.Exit(1)
    }
    cpb, cr := pem.Decode(cf)
    fmt.Println(string(cr))
    kpb, kr := pem.Decode(kf)
    fmt.Println(string(kr))
    crt, e := x509.ParseCertificate(cpb.Bytes)

    if e != nil {
        fmt.Println("parsex509:", e.Error())
        os.Exit(1)
    }
    key, e := x509.ParsePKCS1PrivateKey(kpb.Bytes)
    if e != nil {
        fmt.Println("parsekey:", e.Error())
        os.Exit(1)
    }
    return crt, key
}
*/

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the serial number of the certificate
func (c *Cert) Serial() string {
	return c.serial.String()
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
func generateKey(t KeyType) (any, any, error) {
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
