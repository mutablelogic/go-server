// Package config advertises a server's configuration, at /config.json, to its
// frontend: the path prefix of its API, its name, its version, when it was
// built, where the API's documentation is, and its public endpoint, when it
// has one
package config

import (
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	jsonschema "github.com/mutablelogic/go-server/pkg/jsonschema"
	types "github.com/mutablelogic/go-server/pkg/types"
	version "github.com/mutablelogic/go-server/pkg/version"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Config is a server's configuration, which its frontend reads from
// /config.json, so it knows where the API is
type Config struct {
	Prefix  string `json:"prefix" help:"Path prefix of the API, such as /api"`
	Name    string `json:"name,omitempty" help:"Name of the server, such as its command"`
	Version string `json:"version,omitempty" help:"Version of the server"`

	// When the server was built, so its frontend knows it's been rebuilt,
	// and offers to refresh the page, even when its version is the same,
	// such as while it's developed
	Build string `json:"build,omitempty" help:"When the server was built" format:"date-time"`
	OpenAPI string `json:"openapi,omitempty" help:"Path of the API's documentation, when it's served"`

	// The server's public endpoint, such as https://example.com/api, when it
	// was given its public name, such as that of a proxy it's behind, with
	// --tls.name, so it's https, or none, when it wasn't, as its own URL is
	// then its listen address, which its frontend can't always reach
	Endpoint string `json:"endpoint,omitempty" help:"The server's public endpoint, when it was given its public name, such as https://example.com/api" format:"uri"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// Path is where the configuration is served, which is beside the
	// frontend, rather than beneath the API's prefix, which it says
	Path = "/config.json"

	// Tag is the configuration's tag, which a server with tag groups adds to
	// one of them, as Redoc doesn't show a tag which isn't in a group
	Tag = "Config"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New returns the configuration of the server which the command runs, with
// the router's prefix, the command's name and version, when it was built,
// when it's known, the path of the
// API's documentation, when openapi is true, as it's served, and the
// command's URL, as its public endpoint, when it's https. It's called once
// the server's URL is known, such as when its routes are registered.
func New(ctx server.Cmd, router *httprouter.Router, openapi bool) Config {
	config := Config{
		Prefix:   router.Prefix(),
		Name:     ctx.Name(),
		Version:  ctx.Version(),
		Endpoint: PublicEndpoint(ctx.URL()),
	}
	if built := version.BuildTime(); !built.IsZero() {
		config.Build = built.Format(time.RFC3339)
	}
	if openapi {
		config.OpenAPI = path.Join(router.Prefix(), "openapi.html")
	}
	return config
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// RegisterHandler registers the configuration with the router, at Path,
// with Tag
func RegisterHandler(router *httprouter.Router, config Config) error {
	return router.Register(Path, nil, func(path httprequest.PathItem) {
		path.Tag(Tag)

		// GET
		path.Get(func(w http.ResponseWriter, r *http.Request) {
			httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), config)
		}, func(op httprequest.PathOperation) {
			op.Summary("Get Configuration")
			op.Description("The server's configuration, for its frontend: the path prefix of the API, the server's name and version, the path of the API's documentation, when it's served, and the server's public endpoint, when it was given its public name.")
			op.JSONResponse(http.StatusOK, jsonschema.MustFor[Config](), "Configuration")
		})
	})
}

// PublicEndpoint returns the URL, when it's https, as it is when the server
// was given its public name, without a slash at its end, or none
func PublicEndpoint(endpoint *url.URL) string {
	if endpoint == nil || endpoint.Scheme != types.SchemeSecure || endpoint.Host == "" {
		return ""
	}
	// Its path is escaped, so an escaped character, such as %23, stays one
	return strings.TrimSuffix(endpoint.Scheme+"://"+endpoint.Host+endpoint.EscapedPath(), "/")
}
