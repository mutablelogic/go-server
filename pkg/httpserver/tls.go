package httpserver

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	PemTypePrivateKey  = "PRIVATE KEY"
	PemTypeCertificate = "CERTIFICATE"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// TLSConfig creates a [tls.Config] from raw PEM-encoded certificate and
// key data. The data slices may each contain a single PEM block, or they
// may be concatenated (cert + key in one slice). name sets the
// ServerName field; when verify is false, client certificate verification
// is skipped.
func TLSConfig(name string, verify bool, data ...[]byte) (*tls.Config, error) {
	var buf bytes.Buffer
	for _, d := range data {
		if len(d) == 0 {
			continue
		}
		buf.Write(d)
		buf.WriteByte('\n')
	}

	// Read key pair
	cert, key, err := readPemBlocks(bytes.TrimSpace(buf.Bytes()))
	if err != nil {
		return nil, err
	}

	// Load certificate
	certificate, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	// Return success
	return &tls.Config{ServerName: name, InsecureSkipVerify: !verify, Certificates: []tls.Certificate{certificate}}, nil
}

// ValidateCert checks that the given PEM-encoded certificate and key
// form a valid key pair and that the leaf certificate has not expired.
// It is intended for use at plan time to catch configuration errors
// before Apply.
func ValidateCert(cert, key []byte) error {
	certPEM, keyPEM, err := readPemBlocks(bytes.TrimSpace(append(append([]byte{}, cert...), key...)))
	if err != nil {
		return err
	}

	// Verify the key pair is valid
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return fmt.Errorf("invalid key pair: %w", err)
	}

	// Parse the leaf certificate to check expiry
	leaf, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return fmt.Errorf("cannot parse certificate: %w", err)
	}
	if time.Now().After(leaf.NotAfter) {
		return fmt.Errorf("certificate expired on %s", leaf.NotAfter.Format(time.DateOnly))
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func readPemBlocks(data []byte) ([]byte, []byte, error) {
	var cert, key []byte

	// Read certificate and key from buffer
	for len(data) > 0 {
		// Decode the PEM block
		var block *pem.Block
		block, data = pem.Decode(data)
		if block == nil {
			return nil, nil, fmt.Errorf("invalid PEM block")
		}

		// Set certificate or key
		switch block.Type {
		case PemTypePrivateKey:
			key = pem.EncodeToMemory(block)
		case PemTypeCertificate:
			cert = pem.EncodeToMemory(block)
		default:
			return nil, nil, fmt.Errorf("invalid PEM block type: %q", block.Type)
		}

		// Trim space between the blocks
		data = bytes.TrimSpace(data)
	}

	// If we're missing the certificate or key, return an error
	if cert == nil || key == nil {
		return nil, nil, fmt.Errorf("missing certificate or key")
	}

	// Return success
	return cert, key, nil
}
