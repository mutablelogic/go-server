package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Opt is a function which applies options
type Opt func(*Cert) error

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func apply(opts ...Opt) (*Cert, error) {
	// Create new options
	cert := new(Cert)

	// ECDSA, ED25519 and RSA subject keys should have the DigitalSignature
	// KeyUsage bits set in the x509.Certificate template
	cert.x509.KeyUsage = x509.KeyUsageDigitalSignature

	// Set other defaults
	cert.x509.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
	cert.x509.BasicConstraintsValid = true

	// Apply options
	for _, fn := range opts {
		if err := fn(cert); err != nil {
			return nil, err
		}
	}

	// Return success
	return cert, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Set common name
func WithCommonName(name string) Opt {
	return func(o *Cert) error {
		if name != "" {
			o.x509.Subject.CommonName = name
		}
		return nil
	}
}

// Set organization
func WithOrganization(org, unit string) Opt {
	return func(o *Cert) error {
		if org != "" {
			o.x509.Subject.Organization = []string{org}
		}
		if unit != "" {
			o.x509.Subject.OrganizationalUnit = []string{unit}
		}
		return nil
	}
}

// Set country
func WithCountry(country, state, city string) Opt {
	return func(o *Cert) error {
		if country != "" {
			o.x509.Subject.Country = []string{country}
		}
		if state != "" {
			o.x509.Subject.Province = []string{state}
		}
		if city != "" {
			o.x509.Subject.Locality = []string{city}
		}
		return nil
	}
}

// Set address
func WithAddress(address, postcode string) Opt {
	return func(o *Cert) error {
		if address != "" {
			o.x509.Subject.StreetAddress = []string{address}
		}
		if postcode != "" {
			o.x509.Subject.PostalCode = []string{postcode}
		}
		return nil
	}
}

// Set certificate expiry
func WithExpiry(expires time.Duration) Opt {
	return func(o *Cert) error {
		o.x509.NotBefore = time.Now()
		o.x509.NotAfter = o.x509.NotBefore.Add(expires)
		return nil
	}
}

// Set random serial number
func WithRandomSerial() Opt {
	return func(o *Cert) error {
		// Generate a random serial number
		serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
		if serialNumber, err := rand.Int(rand.Reader, serialNumberLimit); err != nil {
			return err
		} else {
			o.x509.SerialNumber = serialNumber
		}

		// Return success
		return nil
	}
}

// Set serial number
func WithSerial(serial *big.Int) Opt {
	return func(o *Cert) error {
		if serial == nil {
			return WithRandomSerial()(o)
		} else {
			o.x509.SerialNumber = serial
		}
		return nil
	}
}

// Create an ECDSA key with one of the following curves: P224, P256, P384, P521
func WithEllipticKey(t string) Opt {
	return func(o *Cert) error {
		// Generate a private key
		if key, err := ecdsaKey(t); err != nil {
			return err
		} else {
			o.priv = key
		}

		// Return success
		return nil
	}
}

// Create an RSA key with the specified number of bits
func WithRSAKey(bits int) Opt {
	return func(o *Cert) error {
		// Set bits if not specified
		if bits <= 0 {
			bits = defaultBits
		}
		// Generate a private key
		if key, err := rsa.GenerateKey(rand.Reader, bits); err != nil {
			return err
		} else {
			o.priv = key
		}

		// RSA subject keys should have the KeyEncipherment KeyUsage bits set
		o.x509.KeyUsage |= x509.KeyUsageKeyEncipherment

		// Return success
		return nil
	}
}

// Set hosts and IP addreses for the certificate
func WithAddr(addr ...string) Opt {
	return func(o *Cert) error {
		for _, addr := range addr {
			addr = strings.TrimSpace(addr)
			if ip := net.ParseIP(addr); ip != nil {
				o.x509.IPAddresses = append(o.x509.IPAddresses, ip)
			} else {
				o.x509.DNSNames = append(o.x509.DNSNames, addr)
			}
		}

		// Return success
		return nil
	}
}

// Set as a CA certificate
func WithCA() Opt {
	return func(o *Cert) error {
		o.x509.IsCA = true
		o.x509.KeyUsage |= x509.KeyUsageCertSign
		return nil
	}
}

// Set the signer for the certificate
func WithSigner(signer *Cert) Opt {
	return func(o *Cert) error {
		if signer == nil {
			return fmt.Errorf("missing signer")
		} else if !signer.IsCA() {
			return fmt.Errorf("signer is not a CA certificate")
		}
		o.Signer = signer
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func ecdsaKey(t string) (*ecdsa.PrivateKey, error) {
	switch strings.ToUpper(t) {
	case "P224":
		return ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "P256":
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P384":
		return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "P521":
		return ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		return nil, fmt.Errorf("unrecognized key type: %q", t)
	}
}
