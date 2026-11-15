// Package config advertises a server's configuration, at /config.json, to its
// frontend: the path prefix of its API, its name, its version, when it was
// built, where the API's documentation is, and its public endpoint, when it
// has one
package config

import (
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"
	"time"

	// Packages
	upstream "github.com/google/jsonschema-go/jsonschema"
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
	Build   string `json:"build,omitempty" help:"When the server was built" format:"date-time"`
	OpenAPI string `json:"openapi,omitempty" help:"Path of the API's documentation, when it's served"`

	// The server's public endpoint, such as https://example.com/api, when it
	// was given its public name, such as that of a proxy it's behind, with
	// --tls.name, so it's https, or none, when it wasn't, as its own URL is
	// then its listen address, which its frontend can't always reach
	Endpoint string `json:"endpoint,omitempty" help:"The server's public endpoint, when it was given its public name, such as https://example.com/api" format:"uri"`

	// What the server's modules, such as its packages, say of themselves,
	// by their names, such as what they can do, which RegisterHandler sets
	// from WithSection
	Modules map[string]any `json:"modules,omitempty" help:"What the server's modules say of themselves, by their names, such as what they can do"`
}

// Opt is an option of the configuration's handler, such as a module's
// section
type Opt func(*opts) error

type opts struct {
	sections []section
}

// section is a module's section of the configuration, and its schema
type section struct {
	name   string
	value  any
	schema *jsonschema.Schema
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

// WithSection adds a module's section to the configuration, under modules,
// by its name, such as credential, with its value, such as what the module
// can do, which the API's documentation describes, by its type
func WithSection[T any](name string, value T) Opt {
	return func(o *opts) error {
		if strings.TrimSpace(name) == "" {
			return errors.New("config: a section's name is required")
		} else if slices.ContainsFunc(o.sections, func(s section) bool { return s.name == name }) {
			return fmt.Errorf("config: section %q is added more than once", name)
		}
		schema, err := jsonschema.For[T]()
		if err != nil {
			return fmt.Errorf("config: section %q: %w", name, err)
		}
		o.sections = append(o.sections, section{name: name, value: value, schema: schema})
		return nil
	}
}

// RegisterHandler registers the configuration with the router, at Path,
// with Tag, and the modules' sections, which are documented by their types
func RegisterHandler(router *httprouter.Router, config Config, options ...Opt) error {
	var o opts
	for _, opt := range options {
		if err := opt(&o); err != nil {
			return err
		}
	}

	// The modules' sections, which replace what the configuration has
	if len(o.sections) > 0 {
		if len(config.Modules) > 0 {
			return errors.New("config: modules are set with WithSection, not in the configuration")
		}
		config.Modules = make(map[string]any, len(o.sections))
		for _, section := range o.sections {
			config.Modules[section.name] = section.value
		}
	}

	return router.Register(Path, nil, func(path httprequest.PathItem) {
		path.Tag(Tag)

		// GET
		path.Get(func(w http.ResponseWriter, r *http.Request) {
			httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), config)
		}, func(op httprequest.PathOperation) {
			op.Summary("Get Configuration")
			op.Description("The server's configuration, for its frontend: the path prefix of the API, the server's name and version, the path of the API's documentation, when it's served, the server's public endpoint, when it was given its public name, and what its modules say of themselves.")
			op.JSONResponse(http.StatusOK, responseSchema(o.sections), "Configuration")
		})
	})
}

// responseSchema returns the configuration's schema, whose modules are the
// sections, by their names, each with its own schema. The schema of Config
// is cached, so it's copied, rather than changed.
func responseSchema(sections []section) *jsonschema.Schema {
	base := jsonschema.MustFor[Config]()
	if len(sections) == 0 {
		return base
	}
	modules := &upstream.Schema{
		Type:       "object",
		Properties: make(map[string]*upstream.Schema, len(sections)),
	}
	if property := base.Properties["modules"]; property != nil {
		modules.Description = property.Description
	}
	for _, section := range sections {
		modules.Properties[section.name] = &section.schema.Schema
		modules.Required = append(modules.Required, section.name)
	}
	schema := *base
	schema.Properties = maps.Clone(base.Properties)
	schema.Properties["modules"] = modules
	return &schema
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
