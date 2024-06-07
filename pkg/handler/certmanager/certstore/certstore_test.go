package certstore_test

import (
	"os/user"
	"path/filepath"
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/handler/certmanager/cert"
	"github.com/mutablelogic/go-server/pkg/handler/certmanager/certstore"
	"github.com/stretchr/testify/assert"
)

func Test_certstore_001(t *testing.T) {
	assert := assert.New(t)

	// What is my group
	u, err := user.Current()
	if !assert.NoError(err) {
		t.SkipNow()
	}

	store, err := certstore.New(certstore.Config{
		DataPath: filepath.Join(t.TempDir(), "a/b/c"),
		Group:    u.Gid,
	})
	if !assert.NoError(err) {
		t.SkipNow()
	}
	assert.NotNil(store)
}

func Test_certstore_002(t *testing.T) {
	assert := assert.New(t)

	store, err := certstore.New(certstore.Config{
		DataPath: t.TempDir(),
	})
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// Create a new certificate
	cert, err := cert.NewCA(t.Name())
	if !assert.NoError(err) {
		t.SkipNow()
	}

	t.Log(cert)

	// Store the CA
	if err := store.Write(cert); !assert.NoError(err) {
		t.SkipNow()
	}

	// Read the CA
	if _, err := store.Read(cert.Serial()); !assert.NoError(err) {
		t.SkipNow()
	}

	// Delete the CA
	if err := store.Delete(cert); !assert.NoError(err) {
		t.SkipNow()
	}
}

func Test_certstore_003(t *testing.T) {
	assert := assert.New(t)

	// What is my group
	u, err := user.Current()
	if !assert.NoError(err) {
		t.SkipNow()
	}

	store, err := certstore.New(certstore.Config{
		DataPath: t.TempDir(),
		Group:    u.Gid,
	})
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// Create 10 CA's
	for i := 0; i < 10; i++ {
		cert, err := cert.NewCA(t.Name())
		if !assert.NoError(err) {
			t.SkipNow()
		}
		if err := store.Write(cert); !assert.NoError(err) {
			t.SkipNow()
		}
	}

	// List all CA's
	certs, err := store.List()
	if !assert.NoError(err) {
		t.SkipNow()
	}

	t.Log(certs)
}
