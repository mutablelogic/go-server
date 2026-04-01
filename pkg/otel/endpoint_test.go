package otel

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func TestParseEndpoint_URLPreservesSchemeAndHost(t *testing.T) {
	assert := assert.New(t)

	parsed, err := parseEndpoint("https://loki.mutablelogic.com/otlp")
	if !assert.NoError(err) {
		return
	}

	assert.Equal("https", parsed.Scheme)
	assert.Equal("loki.mutablelogic.com", parsed.Host)
	assert.Equal("/otlp", parsed.Path)
}

func TestParseEndpoint_HostPortGetsHTTPSPrefix(t *testing.T) {
	assert := assert.New(t)

	parsed, err := parseEndpoint("localhost:4318")
	if !assert.NoError(err) {
		return
	}

	assert.Equal("https", parsed.Scheme)
	assert.Equal("localhost:4318", parsed.Host)
}
