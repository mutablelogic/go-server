package schema

import (
	"encoding/json"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Spec struct {
	Openapi           string   `json:"openapi"`
	Info              Info     `json:"info"`                        // Required.
	JSONSchemaDialect *string  `json:"jsonSchemaDialect,omitempty"` // Format: uri.
	Servers           []Server `json:"servers,omitempty"`
	Paths             *Paths   `json:"paths,omitempty"`
}

type Info struct {
	Title       string  `json:"title"` // Required.
	Summary     *string `json:"summary,omitempty"`
	Description *string `json:"description,omitempty"`
	Version     string  `json:"version"` // Required.
}

type Server struct {
	URL         string  `json:"url"` // Required. Format: uri-reference.
	Description *string `json:"description,omitempty"`
}

type Paths struct {
	MapOfPathItemValues map[string]PathItem `json:"-"` // Key must match pattern: `^/`.
}

type PathItem struct {
	Summary     *string    `json:"summary,omitempty"`
	Description *string    `json:"description,omitempty"`
	Get         *Operation `json:"get,omitempty"`
	Put         *Operation `json:"put,omitempty"`
	Post        *Operation `json:"post,omitempty"`
	Delete      *Operation `json:"delete,omitempty"`
	Options     *Operation `json:"options,omitempty"`
	Head        *Operation `json:"head,omitempty"`
	Patch       *Operation `json:"patch,omitempty"`
	Trace       *Operation `json:"trace,omitempty"`
}

type Operation struct {
	Tags        []string `json:"tags,omitempty"`
	Summary     *string  `json:"summary,omitempty"`
	Description *string  `json:"description,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	Version = "3.1.1" // https://swagger.io/specification/
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

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
