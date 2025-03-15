package httpserver

import (
	"bytes"
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"io"
	"os"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	PemTypePrivateKey  = "PRIVATE KEY"
	PemTypeCertificate = "CERTIFICATE"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new TLS configuration object from the certificate and key files
func TLSConfig(name string, verify bool, files ...string) (*tls.Config, error) {
	var data bytes.Buffer

	// Append certificate and key to buffer
	for _, file := range files {
		if file == "" {
			continue
		}
		if _, err := os.Stat(file); err != nil {
			return nil, fmt.Errorf("file not found: %s", file)
		}
		if err := appendBytes(&data, file); err != nil {
			return nil, err
		}
	}

	// Read key pair
	cert, key, err := readPemBlocks(bytes.TrimSpace(data.Bytes()))
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

func appendBytes(w io.Writer, file string) error {
	// Append file to writer
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	} else if _, err := w.Write(data); err != nil {
		return err
	}
	return nil
}
