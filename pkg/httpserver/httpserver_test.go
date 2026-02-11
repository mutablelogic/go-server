package httpserver_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"testing"
	"time"

	// Packages
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
	assert "github.com/stretchr/testify/assert"
)

///////////////////////////////////////////////////////////////////////////////
// HELPERS

// selfSignedPEM generates a self-signed cert + key in PEM format for
// testing. The certificate is valid for 1 hour for localhost / 127.0.0.1.
func selfSignedPEM(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})

	kb, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: kb})
	return
}

// expiredPEM generates a self-signed cert that expired yesterday.
func expiredPEM(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "expired"},
		NotBefore:    time.Now().Add(-48 * time.Hour),
		NotAfter:     time.Now().Add(-24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	kb, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: kb})
	return
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - NEW

func Test_Server_001(t *testing.T) {
	assert := assert.New(t)

	t.Run("1", func(t *testing.T) {
		server, err := httpserver.New(":8080", nil, nil)
		if assert.NoError(err) {
			assert.NotNil(server)
		}
	})
	t.Run("2", func(t *testing.T) {
		server, err := httpserver.New("", nil, nil)
		if assert.NoError(err) {
			assert.NotNil(server)
		}
	})
	t.Run("3", func(t *testing.T) {
		server, err := httpserver.New("", nil, &tls.Config{})
		if assert.NoError(err) {
			assert.NotNil(server)
		}
	})
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - LISTEN ADDR

func Test_ListenAddr_001(t *testing.T) {
	assert := assert.New(t)

	// Empty string with no TLS → localhost:http
	addr := httpserver.ListenAddr("", false)
	assert.Equal("localhost:http", addr)
}

func Test_ListenAddr_002(t *testing.T) {
	assert := assert.New(t)

	// Empty string with TLS → localhost:https
	addr := httpserver.ListenAddr("", true)
	assert.Equal("localhost:https", addr)
}

func Test_ListenAddr_003(t *testing.T) {
	assert := assert.New(t)

	// Host without port → default port appended
	addr := httpserver.ListenAddr("myhost", false)
	assert.Equal("myhost:http", addr)
}

func Test_ListenAddr_004(t *testing.T) {
	assert := assert.New(t)

	// Host with port → used as-is
	addr := httpserver.ListenAddr("myhost:9090", false)
	assert.Equal("myhost:9090", addr)
}

func Test_ListenAddr_005(t *testing.T) {
	assert := assert.New(t)

	// Just :port → kept as-is
	addr := httpserver.ListenAddr(":3000", false)
	assert.Equal(":3000", addr)
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - OPTIONS

func Test_Options_001(t *testing.T) {
	assert := assert.New(t)

	// WithReadTimeout applies
	s, err := httpserver.New(":0", nil, nil, httpserver.WithReadTimeout(10*time.Second))
	assert.NoError(err)
	assert.NotNil(s)
}

func Test_Options_002(t *testing.T) {
	assert := assert.New(t)

	// WithWriteTimeout applies
	s, err := httpserver.New(":0", nil, nil, httpserver.WithWriteTimeout(15*time.Second))
	assert.NoError(err)
	assert.NotNil(s)
}

func Test_Options_003(t *testing.T) {
	assert := assert.New(t)

	// WithIdleTimeout applies
	s, err := httpserver.New(":0", nil, nil, httpserver.WithIdleTimeout(20*time.Second))
	assert.NoError(err)
	assert.NotNil(s)
}

func Test_Options_004(t *testing.T) {
	assert := assert.New(t)

	// Zero-value durations do not override defaults (no panic)
	s, err := httpserver.New(":0", nil, nil,
		httpserver.WithReadTimeout(0),
		httpserver.WithWriteTimeout(0),
		httpserver.WithIdleTimeout(0),
	)
	assert.NoError(err)
	assert.NotNil(s)
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - TLS CONFIG

func Test_TLSConfig_001(t *testing.T) {
	assert := assert.New(t)

	// Valid cert + key produces a tls.Config
	certPEM, keyPEM := selfSignedPEM(t)
	cfg, err := httpserver.TLSConfig("localhost", false, certPEM, keyPEM)
	assert.NoError(err)
	assert.NotNil(cfg)
	assert.Equal("localhost", cfg.ServerName)
	assert.Len(cfg.Certificates, 1)
}

func Test_TLSConfig_002(t *testing.T) {
	assert := assert.New(t)

	// Cert + key concatenated in a single slice
	certPEM, keyPEM := selfSignedPEM(t)
	combined := append(certPEM, keyPEM...)
	cfg, err := httpserver.TLSConfig("localhost", false, combined)
	assert.NoError(err)
	assert.NotNil(cfg)
	assert.Len(cfg.Certificates, 1)
}

func Test_TLSConfig_003(t *testing.T) {
	assert := assert.New(t)

	// Empty data returns error
	_, err := httpserver.TLSConfig("localhost", false, []byte{})
	assert.Error(err)
}

func Test_TLSConfig_004(t *testing.T) {
	assert := assert.New(t)

	// Invalid PEM returns error
	_, err := httpserver.TLSConfig("localhost", false, []byte("not pem data"))
	assert.Error(err)
}

func Test_TLSConfig_005(t *testing.T) {
	assert := assert.New(t)

	// Only cert, no key returns error
	certPEM, _ := selfSignedPEM(t)
	_, err := httpserver.TLSConfig("localhost", false, certPEM)
	assert.Error(err)
}

func Test_TLSConfig_006(t *testing.T) {
	assert := assert.New(t)

	// Only key, no cert returns error
	_, keyPEM := selfSignedPEM(t)
	_, err := httpserver.TLSConfig("localhost", false, keyPEM)
	assert.Error(err)
}

func Test_TLSConfig_007(t *testing.T) {
	assert := assert.New(t)

	// verify=true sets InsecureSkipVerify=false
	certPEM, keyPEM := selfSignedPEM(t)
	cfg, err := httpserver.TLSConfig("localhost", true, certPEM, keyPEM)
	assert.NoError(err)
	assert.False(cfg.InsecureSkipVerify)
}

func Test_TLSConfig_008(t *testing.T) {
	assert := assert.New(t)

	// verify=false sets InsecureSkipVerify=true
	certPEM, keyPEM := selfSignedPEM(t)
	cfg, err := httpserver.TLSConfig("localhost", false, certPEM, keyPEM)
	assert.NoError(err)
	assert.True(cfg.InsecureSkipVerify)
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - VALIDATE CERT

func Test_ValidateCert_001(t *testing.T) {
	assert := assert.New(t)

	// Valid non-expired cert passes
	certPEM, keyPEM := selfSignedPEM(t)
	assert.NoError(httpserver.ValidateCert(certPEM, keyPEM))
}

func Test_ValidateCert_002(t *testing.T) {
	assert := assert.New(t)

	// Expired cert fails
	certPEM, keyPEM := expiredPEM(t)
	err := httpserver.ValidateCert(certPEM, keyPEM)
	assert.Error(err)
	assert.Contains(err.Error(), "expired")
}

func Test_ValidateCert_003(t *testing.T) {
	assert := assert.New(t)

	// Mismatched cert/key fails
	certPEM1, _ := selfSignedPEM(t)
	_, keyPEM2 := selfSignedPEM(t)
	err := httpserver.ValidateCert(certPEM1, keyPEM2)
	assert.Error(err)
}

func Test_ValidateCert_004(t *testing.T) {
	assert := assert.New(t)

	// Empty input fails
	err := httpserver.ValidateCert([]byte{}, []byte{})
	assert.Error(err)
}

func Test_ValidateCert_005(t *testing.T) {
	assert := assert.New(t)

	// Invalid PEM fails
	err := httpserver.ValidateCert([]byte("garbage"), []byte("garbage"))
	assert.Error(err)
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - LISTEN + ADDR

func Test_Listen_001(t *testing.T) {
	assert := assert.New(t)

	// Listen on :0 succeeds and Addr returns a real port
	s, err := httpserver.New(":0", nil, nil)
	assert.NoError(err)
	assert.NoError(s.Listen())

	addr := s.Addr()
	assert.NotEmpty(addr)
	assert.NotEqual(":0", addr, "expected resolved port")

	// The address should be parseable
	_, port, err := net.SplitHostPort(addr)
	assert.NoError(err)
	assert.NotEqual("0", port)
}

func Test_Listen_002(t *testing.T) {
	assert := assert.New(t)

	// Before Listen, Addr returns the configured address
	s, err := httpserver.New("localhost:9999", nil, nil)
	assert.NoError(err)
	assert.Equal("localhost:9999", s.Addr())
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - RUN

func Test_Run_001(t *testing.T) {
	assert := assert.New(t)

	// Run starts serving, responds to requests, and shuts down on cancel
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	s, err := httpserver.New(":0", mux, nil)
	assert.NoError(err)
	assert.NoError(s.Listen())

	ctx, cancel := context.WithCancel(context.Background())

	// Run in background
	done := make(chan error, 1)
	go func() {
		done <- s.Run(ctx)
	}()

	// Give the server a moment to start
	time.Sleep(50 * time.Millisecond)

	// Make a request
	resp, err := http.Get("http://" + s.Addr() + "/ping")
	assert.NoError(err)
	assert.Equal(http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Cancel and wait for clean shutdown
	cancel()
	err = <-done
	assert.NoError(err)
}

func Test_Run_002(t *testing.T) {
	assert := assert.New(t)

	// Run with TLS
	certPEM, keyPEM := selfSignedPEM(t)
	tlsCfg, err := httpserver.TLSConfig("localhost", false, certPEM, keyPEM)
	assert.NoError(err)

	mux := http.NewServeMux()
	mux.HandleFunc("/secure", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	s, err := httpserver.New(":0", mux, tlsCfg)
	assert.NoError(err)
	assert.NoError(s.Listen())

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- s.Run(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	// Use a client that skips cert verification
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Get("https://" + s.Addr() + "/secure")
	assert.NoError(err)
	assert.Equal(http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	cancel()
	err = <-done
	assert.NoError(err)
}

func Test_Run_003(t *testing.T) {
	assert := assert.New(t)

	// Immediate cancel still returns cleanly
	s, err := httpserver.New(":0", nil, nil)
	assert.NoError(err)
	assert.NoError(s.Listen())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before Run

	err = s.Run(ctx)
	assert.NoError(err)
}
