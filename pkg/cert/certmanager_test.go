package cert_test

import (
	"context"
	"testing"
	"time"

	// Packages
	test "github.com/djthorpe/go-pg/pkg/test"
	cert "github.com/mutablelogic/go-server/pkg/cert"
	schema "github.com/mutablelogic/go-server/pkg/cert/schema"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
	assert "github.com/stretchr/testify/assert"
)

// Global connection variable
var conn test.Conn

// Start up a container and test the pool
func TestMain(m *testing.M) {
	test.Main(m, &conn)
}

/////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Test_CertManager_001(t *testing.T) {
	assert := assert.New(t)
	conn := conn.Begin(t)
	defer conn.Close()

	// Create a new certificate manager with a root certificate
	certmanager, err := cert.NewCertManager(context.TODO(), conn,
		cert.WithCommonName("root"),
		cert.WithRSAKey(4096),
		cert.WithExpiry(16*365*24*time.Hour),
		cert.WithCA(),
	)
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(certmanager)

	// Get the root name
	t.Run("GetName", func(t *testing.T) {
		name, err := certmanager.GetName(context.TODO(), types.PtrUint64(certmanager.Root().Subject))
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(name)
		assert.Equal(name.CommonName, "root")
	})

	// List names
	t.Run("ListName", func(t *testing.T) {
		names, err := certmanager.ListNames(context.TODO())
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(names)
		assert.Equal(names.Count, uint64(1))
		assert.Len(names.Body, 1)
	})

	// Update name
	t.Run("UpdateName", func(t *testing.T) {
		name, err := certmanager.UpdateName(context.TODO(), types.PtrUint64(certmanager.Root().Subject), schema.NameMeta{
			Org: types.StringPtr("mutablelogic"),
		})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal("mutablelogic", types.PtrString(name.Org))
	})

	// Delete name - don't allow the deletion of the root certificate
	t.Run("DeleteName", func(t *testing.T) {
		_, err := certmanager.DeleteName(context.TODO(), types.PtrUint64(certmanager.Root().Subject))
		assert.ErrorIs(err, httpresponse.ErrConflict)
	})
}

func Test_CertManager_002(t *testing.T) {
	assert := assert.New(t)
	conn := conn.Begin(t)
	defer conn.Close()

	// Create a new certificate manager with a root certificate
	certmanager, err := cert.NewCertManager(context.TODO(), conn,
		cert.WithCommonName("root"),
		cert.WithRSAKey(4096),
		cert.WithExpiry(16*365*24*time.Hour),
		cert.WithCA(),
	)
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(certmanager)

	// Get the root cert
	t.Run("GetRootCert", func(t *testing.T) {
		cert, err := certmanager.GetCert(context.TODO(), "root")
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(cert)
		assert.Equal(cert.Name, "root")
		assert.True(cert.IsCA)
	})

	// Root cert cannot be updated
	t.Run("UpdateRootCert", func(t *testing.T) {
		_, err := certmanager.UpdateCert(context.TODO(), "root", schema.CertMeta{})
		assert.ErrorIs(err, httpresponse.ErrConflict)
	})

	// Root cert cannot be deleted
	t.Run("DeleteRootCert", func(t *testing.T) {
		_, err := certmanager.DeleteCert(context.TODO(), "root")
		assert.ErrorIs(err, httpresponse.ErrConflict)
	})

}

func Test_CertManager_003(t *testing.T) {
	assert := assert.New(t)
	conn := conn.Begin(t)
	defer conn.Close()

	// Create a new certificate manager with a root certificate
	certmanager, err := cert.NewCertManager(context.TODO(), conn,
		cert.WithCommonName("root"),
		cert.WithRSAKey(4096),
		cert.WithExpiry(16*365*24*time.Hour),
		cert.WithCA(),
	)
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(certmanager)

	// Create a certificate which is signed by the root certifificate
	t.Run("CreateCert", func(t *testing.T) {
		cert, err := certmanager.CreateCert(context.TODO(), "create_cert_1",
			cert.WithSigner(certmanager.Root()),
			cert.WithCommonName("*.hostname.com"),
			cert.WithRSAKey(0),
			cert.WithExpiry(1*time.Hour),
		)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(cert)

		t.Log(cert)
	})

	// Create a certificate and then try and create again
	t.Run("CreateCertDuplicate", func(t *testing.T) {
		_, err := certmanager.CreateCert(context.TODO(), "create_cert_2",
			cert.WithSigner(certmanager.Root()),
			cert.WithCommonName("*.hostname.com"),
			cert.WithRSAKey(0),
			cert.WithExpiry(1*time.Hour),
		)
		if !assert.NoError(err) {
			t.FailNow()
		}

		_, err = certmanager.CreateCert(context.TODO(), "create_cert_2",
			cert.WithSigner(certmanager.Root()),
			cert.WithCommonName("*.hostname.com"),
			cert.WithRSAKey(0),
			cert.WithExpiry(1*time.Hour),
		)
		assert.ErrorIs(err, httpresponse.ErrConflict)

	})

	// Create a certificate and then get it
	t.Run("GetCert", func(t *testing.T) {
		certa, err := certmanager.CreateCert(context.TODO(), "create_cert_3",
			cert.WithSigner(certmanager.Root()),
			cert.WithCommonName("*.hostname.com"),
			cert.WithRSAKey(0),
			cert.WithExpiry(1*time.Hour),
		)
		if !assert.NoError(err) {
			t.FailNow()
		}

		certb, err := certmanager.GetCert(context.TODO(), "create_cert_3")
		if assert.NoError(err) {
			assert.Equal(certa, certb)
		}

		t.Log(certb)
	})
}
