package cert_test

import (
	"context"
	"testing"
	"time"

	// Packages
	test "github.com/djthorpe/go-pg/pkg/test"
	cert "github.com/mutablelogic/go-server/pkg/cert"
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
}
