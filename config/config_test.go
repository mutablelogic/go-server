package config_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"

	// Packages
	config "github.com/mutablelogic/go-server/config"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	test "github.com/mutablelogic/go-server/test"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

func Test_RegisterHandler(t *testing.T) {
	// The configuration is served at /config.json, beside the API rather than
	// beneath its prefix, and conforms to its documentation
	want := config.Config{Prefix: "/api", Name: "server", Version: "1.2.3", Build: "2026-11-12T18:00:00Z", OpenAPI: "/api/openapi.html", Endpoint: "https://example.com/api"}
	server := test.NewServer(t, func(router *httprouter.Router) error {
		return config.RegisterHandler(router, want)
	})
	resp, err := http.Get(server.URL + config.Path)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var got config.Config
	require.NoError(t, json.Unmarshal(body, &got))
	assert.Equal(t, want, got)
	assert.Empty(t, server.Verify())
}

func Test_PublicEndpoint(t *testing.T) {
	for in, want := range map[string]string{
		"https://example.com/api":       "https://example.com/api",
		"https://example.com/api/":      "https://example.com/api",
		"https://example.com:8443/a%23": "https://example.com:8443/a%23",
		"http://127.0.0.1:8084/api":     "",
		"https:///api":                  "",
	} {
		u, err := url.Parse(in)
		require.NoError(t, err)
		assert.Equal(t, want, config.PublicEndpoint(u), in)
	}
	assert.Empty(t, config.PublicEndpoint(nil))
}
