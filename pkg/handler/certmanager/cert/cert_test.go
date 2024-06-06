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
		X509Name: certmanager.X509Name{
			OrganizationalUnit: "Test",
			Organization:       "Test",
			Country:            "DE",
		},
	})
	assert.NoError(err)

	t.Log(ca)

	cert, err := cert.NewCert(ca, cert.OptKeyType("P224"))
	assert.NoError(err)
	assert.NotNil(cert)

	t.Log(cert)
}

func Test_Cert_005(t *testing.T) {
	assert := assert.New(t)
	ca, err := cert.NewCA(certmanager.Config{
		X509Name: certmanager.X509Name{
			OrganizationalUnit: "Test",
			Organization:       "Test",
			Country:            "DE",
		},
	})
	assert.NoError(err)

	both := new(bytes.Buffer)
	err = ca.WriteCertificate(both)
	assert.NoError(err)
	err = ca.WritePrivateKey(both)
	assert.NoError(err)

	cert2, err := cert.NewFromBytes(both.Bytes())
	assert.NoError(err)

	t.Log(cert2)
}
