// Package schema provides Go types for the OpenAPI 3.1 specification.
//
// Only the subset of the specification needed by this project is
// implemented: top-level [Spec], [Info], [Server], [Paths],
// [PathItem] and [Operation].
//
// Reference: https://spec.openapis.org/oas/v3.1.1
package schema

import (
	"encoding/json"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Spec is the root object of an OpenAPI 3.1 document.
type Spec struct {
	Openapi           string   `json:"openapi"`
	Info              Info     `json:"info"`                       // Required.
	JSONSchemaDialect string   `json:"jsonSchemaDialect,omitzero"` // Format: uri.
	Servers           []Server `json:"servers,omitempty"`
	Paths             *Paths   `json:"paths,omitempty"`
}

// Info provides metadata about the API.
type Info struct {
	Title       string `json:"title"` // Required.
	Summary     string `json:"summary,omitzero"`
	Description string `json:"description,omitzero"`
	Version     string `json:"version"` // Required.
}

// Server represents a server that hosts the API.
type Server struct {
	URL         string `json:"url"` // Required. Format: uri-reference.
	Description string `json:"description,omitzero"`
}

// Paths holds the relative paths to the individual endpoints and their
// operations. The map keys must begin with a forward slash.
type Paths struct {
	MapOfPathItemValues map[string]PathItem `json:"-"` // Key must match pattern: `^/`.
}

// PathItem describes the operations available on a single path.
type PathItem struct {
	Summary     string     `json:"summary,omitzero"`
	Description string     `json:"description,omitzero"`
	Get         *Operation `json:"get,omitempty"`
	Put         *Operation `json:"put,omitempty"`
	Post        *Operation `json:"post,omitempty"`
	Delete      *Operation `json:"delete,omitempty"`
	Options     *Operation `json:"options,omitempty"`
	Head        *Operation `json:"head,omitempty"`
	Patch       *Operation `json:"patch,omitempty"`
	Trace       *Operation `json:"trace,omitempty"`
}

// Operation describes a single API operation on a path.
type Operation struct {
	Tags        []string `json:"tags,omitempty"`
	Summary     string   `json:"summary,omitzero"`
	Description string   `json:"description,omitzero"`
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// Version is the OpenAPI specification version used by this package.
	Version = "3.1.1" // https://swagger.io/specification/
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewSpec returns a new [Spec] initialised with the given API title,
// version and the current OpenAPI [Version].
func NewSpec(title, version string) *Spec {
	return types.Ptr(Spec{
		Openapi: Version,
		Info: Info{
			Title:   title,
			Version: version,
		},
	})
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

// MarshalJSON serialises Paths as a flat JSON object keyed by path.
func (p Paths) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.MapOfPathItemValues)
}

// UnmarshalJSON deserialises a flat JSON object into Paths.
func (p *Paths) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &p.MapOfPathItemValues)
}

// AddPath registers a [PathItem] under the given path. If item is nil
// the call is a no-op. Adding a path that already exists replaces the
// previous entry.
func (s *Spec) AddPath(path string, item *PathItem) {
	if s.Paths == nil {
		s.Paths = &Paths{
			MapOfPathItemValues: make(map[string]PathItem),
		}
	}
	if item != nil {
		s.Paths.MapOfPathItemValues[path] = *item
	}
}

// SetServers replaces the servers list in the spec.
func (s *Spec) SetServers(servers []Server) {
	s.Servers = servers
}
