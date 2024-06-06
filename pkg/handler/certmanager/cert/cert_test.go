package cert_test

import (
	"bytes"
	"testing"

	"github.com/mutablelogic/go-server/pkg/handler/certmanager"
	"github.com/mutablelogic/go-server/pkg/handler/certmanager/cert"
	"github.com/stretchr/testify/assert"
)

func Test_Cert_001(t *testing.T) {
	assert := assert.New(t)
	for i := 0; i < 100; i++ {
		serial := cert.SerialNumber()
		assert.NotNil(serial)
		t.Log(serial)
	}
}

func Test_Cert_002(t *testing.T) {
	assert := assert.New(t)
	cert, err := cert.NewCA(certmanager.Config{})
	assert.NoError(err)
	assert.NotNil(cert)
	t.Log(cert)
}

func Test_Cert_003(t *testing.T) {
	assert := assert.New(t)
	cert, err := cert.NewCA(certmanager.Config{})
	assert.NoError(err)

	public := new(bytes.Buffer)
	err = cert.WriteCertificate(public)
	assert.NoError(err)
	assert.NotEmpty(public.String())
	t.Log(public.String())

	private := new(bytes.Buffer)
	err = cert.WritePrivateKey(private)
	assert.NoError(err)
	assert.NotEmpty(private.String())
	t.Log(private.String())
}

func Test_Cert_004(t *testing.T) {
	assert := assert.New(t)
	ca, err := cert.NewCA(certmanager.Config{
		Organization: "Test",
	}, cert.OptKeyType("P521"))
	assert.NoError(err)

	cert, err := cert.NewCert(ca, cert.OptKeyType("ED25519"))
	assert.NoError(err)
	assert.NotNil(cert)

	public := new(bytes.Buffer)
	err = cert.WriteCertificate(public)
	assert.NoError(err)
	assert.NotEmpty(public.String())
	t.Log(public.String())

	private := new(bytes.Buffer)
	err = cert.WritePrivateKey(private)
	assert.NoError(err)
	assert.NotEmpty(private.String())
	t.Log(private.String())
}
