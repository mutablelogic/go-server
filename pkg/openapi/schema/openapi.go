// Package schema provides Go types for the OpenAPI 3.1 specification.
//
// Only the subset of the specification needed by this project is
// implemented: top-level [Spec], [Info], [Server], [Paths],
// [PathItem], [Operation], [RequestBody] and [MediaType].
//
// Reference: https://spec.openapis.org/oas/v3.1.1
package schema

import (
	"encoding/json"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	jsonschema "github.com/mutablelogic/go-server/pkg/jsonschema"
	types "github.com/mutablelogic/go-server/pkg/types"
	yaml "gopkg.in/yaml.v3"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Spec is the root object of an OpenAPI 3.1 document.
type Spec struct {
	Openapi           string                `json:"openapi"                       yaml:"openapi"`
	Info              Info                  `json:"info"                          yaml:"info"`                        // Required.
	JSONSchemaDialect string                `json:"jsonSchemaDialect,omitzero"    yaml:"jsonSchemaDialect,omitempty"` // Format: uri.
	Servers           []Server              `json:"servers,omitempty"             yaml:"servers,omitempty"`
	Tags              []Tag                 `json:"tags,omitempty"                yaml:"tags,omitempty"`
	Paths             *Paths                `json:"paths,omitempty"               yaml:"paths,omitempty"`
	Components        *Components           `json:"components,omitempty"          yaml:"components,omitempty"`
	Security          []SecurityRequirement `json:"security,omitempty"            yaml:"security,omitempty"`
	TagGroups         []TagGroup            `json:"x-tagGroups,omitempty"         yaml:"x-tagGroups,omitempty"` // Redoc extension.
}

// TagGroup groups tags under a heading in Redoc documentation.
// This is a Redoc vendor extension (x-tagGroups).
type TagGroup struct {
	Name string   `json:"name"             yaml:"name"` // The heading displayed in the sidebar.
	Tags []string `json:"tags"             yaml:"tags"` // Tag names belonging to this group.
}

// Tag adds metadata to a group of operations identified by the same tag name.
type Tag struct {
	Name        string `json:"name"                 yaml:"name"` // Required.
	Description string `json:"description,omitzero" yaml:"description,omitempty"`
}

// Info provides metadata about the API.
type Info struct {
	Title       string `json:"title"                yaml:"title"` // Required.
	Summary     string `json:"summary,omitzero"     yaml:"summary,omitempty"`
	Description string `json:"description,omitzero" yaml:"description,omitempty"`
	Version     string `json:"version"              yaml:"version"` // Required.
}

// Server represents a server that hosts the API.
type Server struct {
	URL         string `json:"url"                  yaml:"url"` // Required. Format: uri-reference.
	Description string `json:"description,omitzero" yaml:"description,omitempty"`
}

// Paths holds the relative paths to the individual endpoints and their
// operations. The map keys must begin with a forward slash.
type Paths struct {
	MapOfPathItemValues map[string]PathItem `json:"-" yaml:"-"` // Key must match pattern: `^/`.
}

// PathItem describes the operations available on a single path.
type PathItem struct {
	Summary     string      `json:"summary,omitzero"     yaml:"summary,omitempty"`
	Description string      `json:"description,omitzero" yaml:"description,omitempty"`
	Parameters  []Parameter `json:"parameters,omitempty"  yaml:"parameters,omitempty"`
	Get         *Operation  `json:"get,omitempty"        yaml:"get,omitempty"`
	Put         *Operation  `json:"put,omitempty"        yaml:"put,omitempty"`
	Post        *Operation  `json:"post,omitempty"       yaml:"post,omitempty"`
	Delete      *Operation  `json:"delete,omitempty"     yaml:"delete,omitempty"`
	Options     *Operation  `json:"options,omitempty"    yaml:"options,omitempty"`
	Head        *Operation  `json:"head,omitempty"       yaml:"head,omitempty"`
	Patch       *Operation  `json:"patch,omitempty"      yaml:"patch,omitempty"`
	Trace       *Operation  `json:"trace,omitempty"      yaml:"trace,omitempty"`
}

// Operation describes a single API operation on a path.
type Operation struct {
	Tags        []string              `json:"tags,omitempty"        yaml:"tags,omitempty"`
	Summary     string                `json:"summary,omitzero"      yaml:"summary,omitempty"`
	Description string                `json:"description,omitzero"  yaml:"description,omitempty"`
	OperationId string                `json:"operationId,omitzero"  yaml:"operationId,omitempty"`
	Deprecated  bool                  `json:"deprecated,omitempty"  yaml:"deprecated,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"  yaml:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses,omitempty"   yaml:"responses,omitempty"`
	Security    []SecurityRequirement `json:"security,omitempty"    yaml:"security,omitempty"`
}

// Response describes a single response from an [Operation].
type Response struct {
	Description string               `json:"description"          yaml:"description"` // Required.
	Content     map[string]MediaType `json:"content,omitempty"    yaml:"content,omitempty"`
}

// Parameter describes a single operation parameter (path, query, header or cookie).
type Parameter struct {
	Name        string             `json:"name"                 yaml:"name"` // Required.
	In          string             `json:"in"                   yaml:"in"`   // Required: "path", "query", "header", "cookie".
	Description string             `json:"description,omitzero" yaml:"description,omitempty"`
	Required    bool               `json:"required,omitempty"   yaml:"required,omitempty"`
	Schema      *jsonschema.Schema `json:"schema,omitempty"     yaml:"schema,omitempty"`
}

// RequestBody describes the body of a request for an [Operation].
type RequestBody struct {
	Description string               `json:"description,omitzero" yaml:"description,omitempty"`
	Required    bool                 `json:"required,omitempty"   yaml:"required,omitempty"`
	Content     map[string]MediaType `json:"content"              yaml:"content"`
}

// MediaType provides the schema for a specific content-type in a [RequestBody].
type MediaType struct {
	Schema *jsonschema.Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// SecurityRequirement maps security scheme names to the scopes required.
// For schemes that do not use scopes (e.g. bearer token), the value is an
// empty slice.
type SecurityRequirement map[string][]string

// Components holds reusable objects for the OpenAPI specification.
type Components struct {
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
}

// SecurityScheme describes an authentication mechanism used by the API.
type SecurityScheme struct {
	Type         string      `json:"type"                    yaml:"type"` // Required: "apiKey", "http", "mutualTLS", "oauth2", "openIdConnect".
	Description  string      `json:"description,omitzero"    yaml:"description,omitempty"`
	Name         string      `json:"name,omitzero"           yaml:"name,omitempty"`               // Required for apiKey.
	In           string      `json:"in,omitzero"             yaml:"in,omitempty"`                 // Required for apiKey: "query", "header", "cookie".
	Scheme       string      `json:"scheme,omitzero"         yaml:"scheme,omitempty"`             // Required for http: e.g. "bearer".
	BearerFormat string      `json:"bearerFormat,omitzero"   yaml:"bearerFormat,omitempty"`       // Optional hint for http/bearer.
	Flows        *OAuthFlows `json:"flows,omitempty"         yaml:"flows,omitempty"`              // Required for oauth2.
	OpenIdURL    string      `json:"openIdConnectUrl,omitzero" yaml:"openIdConnectUrl,omitempty"` // Required for openIdConnect.
}

// OAuthFlows describes the available OAuth2 flows.
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty"          yaml:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"          yaml:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
}

// OAuthFlow describes a single OAuth2 flow.
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitzero" yaml:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitzero"         yaml:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitzero"       yaml:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"                    yaml:"scopes"` // Required.
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// Version is the OpenAPI specification version used by this package.
	Version = "3.1.1" // https://swagger.io/specification/

	// ParameterInPath is used for path parameters such as {id}.
	ParameterInPath = "path"
	// ParameterInQuery is used for query string parameters.
	ParameterInQuery = "query"
	// ParameterInHeader is used for header parameters.
	ParameterInHeader = "header"
	// ParameterInCookie is used for cookie parameters.
	ParameterInCookie = "cookie"

	// SecuritySchemeHTTP is the type value for HTTP authentication (e.g. bearer).
	SecuritySchemeHTTP = "http"
	// SecuritySchemeAPIKey is the type value for API key authentication.
	SecuritySchemeAPIKey = "apiKey"
	// SecuritySchemeOAuth2 is the type value for OAuth 2.0 authentication.
	SecuritySchemeOAuth2 = "oauth2"
	// SecuritySchemeOpenIDConnect is the type value for OpenID Connect.
	SecuritySchemeOpenIDConnect = "openIdConnect"
	// SecuritySchemeMutualTLS is the type value for mutual TLS.
	SecuritySchemeMutualTLS = "mutualTLS"
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

// MarshalYAML serialises MediaType via its JSON form so that the nested
// *jsonschema.Schema is rendered compactly (only non-zero fields, using JSON
// omitempty tags that the third-party type provides).
func (m MediaType) MarshalYAML() (any, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// MarshalYAML serialises Parameter via its JSON form for the same reason as
// MediaType.MarshalYAML.
func (p Parameter) MarshalYAML() (any, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// MarshalJSON serialises Paths as a flat JSON object keyed by path.
func (p Paths) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.MapOfPathItemValues)
}

// UnmarshalJSON deserialises a flat JSON object into Paths.
func (p *Paths) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &p.MapOfPathItemValues)
}

// MarshalYAML serialises Paths as a flat YAML mapping keyed by path.
func (p Paths) MarshalYAML() (any, error) {
	node := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	for k, v := range p.MapOfPathItemValues {
		keyNode := &yaml.Node{}
		if err := keyNode.Encode(k); err != nil {
			return nil, err
		}
		valNode := &yaml.Node{}
		if err := valNode.Encode(v); err != nil {
			return nil, err
		}
		node.Content = append(node.Content, keyNode, valNode)
	}
	return node, nil
}

// UnmarshalYAML deserialises a flat YAML mapping into Paths.
func (p *Paths) UnmarshalYAML(value *yaml.Node) error {
	return value.Decode(&p.MapOfPathItemValues)
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

// SetSummary sets the summary field of the spec's Info object.
func (s *Spec) SetSummary(summary string) {
	s.Info.Summary = summary
}

// AddTag appends a named tag to the spec's top-level tag list, which groups
// operations in documentation tools such as Redoc and Swagger UI. If a tag
// with the same name already exists it is replaced.
func (s *Spec) AddTag(name, description string) {
	for i, t := range s.Tags {
		if t.Name == name {
			s.Tags[i].Description = description
			return
		}
	}
	s.Tags = append(s.Tags, Tag{Name: name, Description: description})
}

// AddTagGroup appends a [TagGroup] that groups the given tags under a sidebar
// heading in Redoc. If a group with the same name already exists, its tags
// are replaced.
func (s *Spec) AddTagGroup(name string, tags ...string) {
	for i, g := range s.TagGroups {
		if g.Name == name {
			s.TagGroups[i].Tags = tags
			return
		}
	}
	s.TagGroups = append(s.TagGroups, TagGroup{Name: name, Tags: tags})
}

// AddSecurityScheme registers a named [SecurityScheme] under components.
// If a scheme with the same name already exists it is replaced.
func (s *Spec) AddSecurityScheme(name string, scheme SecurityScheme) {
	if s.Components == nil {
		s.Components = &Components{
			SecuritySchemes: make(map[string]SecurityScheme),
		}
	}
	if s.Components.SecuritySchemes == nil {
		s.Components.SecuritySchemes = make(map[string]SecurityScheme)
	}
	s.Components.SecuritySchemes[name] = scheme
}

// ErrorResponse returns a [Response] whose schema matches the JSON error body
// returned by all handlers that use [httpresponse.Error]. Use the key
// "default" (catches any undeclared status code) or a specific code like "400".
func ErrorResponse(description string) Response {
	s, _ := jsonschema.For[httpresponse.ErrResponse]()
	return Response{
		Description: description,
		Content: map[string]MediaType{
			types.ContentTypeJSON: {Schema: s},
		},
	}
}
