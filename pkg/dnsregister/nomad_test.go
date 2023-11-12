package dnsregister_test

import (
	"net/url"
	"os"
	"testing"

	// Packages
	client "github.com/mutablelogic/go-server/pkg/client"
	dnsregister "github.com/mutablelogic/go-server/pkg/dnsregister"
	assert "github.com/stretchr/testify/assert"
)

func Test_Nomad_000(t *testing.T) {
	assert := assert.New(t)
	uri := uri(t)
	client, err := client.New(client.OptEndpoint(uri.String()))
	assert.NoError(err)
	assert.NotNil(client)
	t.Log(client)
}

func Test_Nomad_001(t *testing.T) {
	assert := assert.New(t)
	uri := uri(t)
	client, err := client.New(client.OptEndpoint(uri.String()), client.OptReqToken(token(t)))
	assert.NoError(err)

	nomad, err := dnsregister.NewNomadClient(client)
	assert.NoError(err)
	assert.NotNil(nomad)

	services, err := nomad.Services()
	assert.NoError(err)
	assert.NotNil(services)
	t.Log(services)

}

func Test_Nomad_002(t *testing.T) {
	assert := assert.New(t)
	uri := uri(t)
	client, err := client.New(client.OptEndpoint(uri.String()), client.OptReqToken(token(t)))
	assert.NoError(err)

	nomad, err := dnsregister.NewNomadClient(client)
	assert.NoError(err)
	assert.NotNil(nomad)

	services, err := nomad.Services()
	assert.NoError(err)

	for _, ns := range services {
		for _, service := range ns.Services {
			service, err := nomad.Service(service.Name)
			assert.NoError(err)
			t.Log(service)
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// Return URL or skip test

func uri(t *testing.T) *url.URL {
	if uri := os.ExpandEnv("${NOMAD_ADDR}"); uri == "" {
		t.Skip("no NOMAD_ADDR environment variable, skipping test")
	} else if uri, err := url.Parse(uri); err != nil {
		t.Skip("invalid MONGO_URL environment variable, skipping test")
	} else {
		return uri
	}
	return nil
}

func token(t *testing.T) client.Token {
	token := os.ExpandEnv("${NOMAD_TOKEN}")
	if token == "" {
		t.Skip("no NOMAD_TOKEN environment variable, skipping test")
		return client.Token{}
	} else {
		return client.Token{Scheme: "Bearer", Value: token}
	}
}
