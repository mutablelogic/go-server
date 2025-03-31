package cert_test

import (
	"bytes"
	"testing"
	"time"

	// Packages
	cert "github.com/mutablelogic/go-server/pkg/cert"
	assert "github.com/stretchr/testify/assert"
)

func Test_Cert_001(t *testing.T) {
	assert := assert.New(t)

	t.Run("1", func(t *testing.T) {
		cert, err := cert.New()
		if assert.Error(err) {
			assert.Nil(cert)
			t.Log(err)
		}
	})

	t.Run("2", func(t *testing.T) {
		cert, err := cert.New(cert.WithRSAKey(0))
		if assert.Error(err) {
			assert.Nil(cert)
			t.Log(err)
		}
	})

	t.Run("3", func(t *testing.T) {
		cert, err := cert.New(
			cert.WithRSAKey(0),
			cert.WithExpiry(time.Hour),
		)
		if assert.NoError(err) {
			assert.NotNil(cert)
			assert.NotNil(cert.PrivateKey())
			assert.NotNil(cert.PublicKey())
		}
	})

}

func Test_Cert_002(t *testing.T) {
	assert := assert.New(t)

	ca, err := cert.New(
		cert.WithCommonName("CA"),
		cert.WithRSAKey(0),
		cert.WithExpiry(time.Hour),
		cert.WithCA(),
	)
	if assert.NoError(err) {
		assert.NotNil(ca)
	}

	t.Run("1", func(t *testing.T) {
		var data bytes.Buffer
		cert, err := cert.New(
			cert.WithRSAKey(0),
			cert.WithExpiry(time.Hour),
			cert.WithSigner(ca),
		)
		if assert.NoError(err) {
			assert.NoError(cert.Write(&data))
			t.Log(data.String())
		}
		if assert.NoError(err) {
			assert.NoError(cert.WritePrivateKey(&data))
			t.Log(data.String())
		}
	})

	t.Run("2", func(t *testing.T) {
		var data bytes.Buffer
		cert1, err := cert.New(
			cert.WithRSAKey(0),
			cert.WithExpiry(time.Hour),
			cert.WithSigner(ca),
		)
		if !assert.NoError(err) {
			t.FailNow()
		}
		// Write the certificate and key
		assert.NoError(cert1.Write(&data))
		assert.NoError(cert1.WritePrivateKey(&data))

		// Read back into another cert
		cert2, err := cert.Read(&data)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(cert2)
		t.Log(cert1.String())
		t.Log(cert2.String())
		assert.Equal(cert1.String(), cert2.String())
	})
}
