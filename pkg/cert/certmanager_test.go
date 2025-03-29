package cert_test

import (
	"context"
	"testing"
	"time"

	// Packages
	test "github.com/djthorpe/go-pg/pkg/test"
	cert "github.com/mutablelogic/go-server/pkg/cert"
	schema "github.com/mutablelogic/go-server/pkg/cert/schema"
	"github.com/mutablelogic/go-server/pkg/types"
	"github.com/mutablelogic/go-service/pkg/httpresponse"
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
			CommonName: "root_2",
			Org:        types.StringPtr("mutablelogic"),
		})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal("root_2", name.CommonName)
		assert.Equal("mutablelogic", types.PtrString(name.Org))
	})

	// Delete name - don't allow the deletion of the root certificate
	t.Run("DeleteName", func(t *testing.T) {
		_, err := certmanager.DeleteName(context.TODO(), types.PtrUint64(certmanager.Root().Subject))
		assert.ErrorIs(err, httpresponse.ErrConflict)
	})
}
