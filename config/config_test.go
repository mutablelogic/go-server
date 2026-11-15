package config_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"

	// Packages
	config "github.com/mutablelogic/go-server/config"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	jsonschema "github.com/mutablelogic/go-server/pkg/jsonschema"
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

// features is a module's section, such as what it can do
type features struct {
	Decrypt bool `json:"decrypt" help:"Whether a value can be read with its passphrase"`
	Rotate  bool `json:"rotate" help:"Whether a value can be rotated to the latest passphrase"`
}

func Test_Sections(t *testing.T) {
	want := config.Config{Prefix: "/api", Name: "server"}
	server := test.NewServer(t, func(router *httprouter.Router) error {
		return config.RegisterHandler(router, want,
			config.WithSection("credential", features{Decrypt: true}),
			config.WithSection("queue", map[string]int{"workers": 4}),
		)
	})

	t.Run("Served", func(t *testing.T) {
		// The sections are under modules, by their names, and conform to
		// the documentation
		resp, err := http.Get(server.URL + config.Path)
		require.NoError(t, err)
		defer resp.Body.Close()
		var got struct {
			config.Config
			Modules struct {
				Credential features       `json:"credential"`
				Queue      map[string]int `json:"queue"`
			} `json:"modules"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
		assert.Equal(t, "server", got.Name)
		assert.Equal(t, features{Decrypt: true}, got.Modules.Credential)
		assert.Equal(t, map[string]int{"workers": 4}, got.Modules.Queue)
		assert.Empty(t, server.Verify())
	})

	t.Run("Documented", func(t *testing.T) {
		// Each section is described by its type, in the response's schema
		schema := server.Spec().Paths.MapOfPathItemValues[config.Path].Get.Responses["200"].Content["application/json"].Schema
		require.NotNil(t, schema)
		modules := schema.Properties["modules"]
		require.NotNil(t, modules)
		assert.ElementsMatch(t, []string{"credential", "queue"}, modules.Required)
		credential := modules.Properties["credential"]
		require.NotNil(t, credential)
		require.Contains(t, credential.Properties, "rotate")
		assert.Equal(t, "Whether a value can be rotated to the latest passphrase", credential.Properties["rotate"].Description)

		// The schema validates the sections by their own schemas, so a
		// section of the wrong type isn't valid
		assert.NoError(t, schema.Validate(json.RawMessage(`{"prefix":"/api","modules":{"credential":{"decrypt":true,"rotate":false},"queue":{"workers":4}}}`)))
		assert.Error(t, schema.Validate(json.RawMessage(`{"prefix":"/api","modules":{"credential":{"decrypt":"yes","rotate":false},"queue":{"workers":4}}}`)), "decrypt is a boolean")
		assert.Error(t, schema.Validate(json.RawMessage(`{"prefix":"/api","modules":{"queue":{"workers":4}}}`)), "credential is required")

		// The schema of Config, which is cached, isn't changed
		base := jsonschema.MustFor[config.Config]().Properties["modules"]
		assert.NotContains(t, base.Properties, "credential")
	})
}

func Test_NoSections(t *testing.T) {
	// Without sections, there are no modules
	server := test.NewServer(t, func(router *httprouter.Router) error {
		return config.RegisterHandler(router, config.Config{Prefix: "/api"})
	})
	resp, err := http.Get(server.URL + config.Path)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.NotContains(t, string(body), "modules")
}

func Test_SectionErrors(t *testing.T) {
	register := func(c config.Config, opts ...config.Opt) error {
		router, err := httprouter.NewRouter(context.Background(), http.NewServeMux(), "/api", "", "test", "0")
		require.NoError(t, err)
		return config.RegisterHandler(router, c, opts...)
	}
	assert.Error(t, register(config.Config{}, config.WithSection("", features{})), "a name is required")
	assert.Error(t, register(config.Config{}, config.WithSection(" ", features{})), "a name is required")
	assert.Error(t, register(config.Config{}, config.WithSection("credential", features{}), config.WithSection("credential", features{})), "a name is unique")
	assert.Error(t, register(config.Config{Modules: map[string]any{"x": 1}}, config.WithSection("credential", features{})), "modules are set by sections")
	assert.NoError(t, register(config.Config{Modules: map[string]any{"x": 1}}), "modules can be set without sections")
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
