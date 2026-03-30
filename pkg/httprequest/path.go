package httprequest

import (
	"net/http"
	"strings"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	jsonschema "github.com/mutablelogic/go-server/pkg/jsonschema"
	openapi "github.com/mutablelogic/go-server/pkg/openapi/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Path struct {
	path       string
	spec       openapi.PathItem
	parameters []openapi.Parameter
	handlers   map[string]http.HandlerFunc
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewPath(path string, summary string) *Path {
	self := new(Path)

	// Path - can have either '/' prefix or not, but will always be normalised
	// to remove '/' from the end
	self.path = types.NormalisePath(path)
	if !strings.HasPrefix(path, "/") && strings.HasPrefix(self.path, "/") {
		self.path = self.path[1:]
	}

	// Set the spec, parameters and handler
	self.spec = openapi.PathItem{
		Summary: summary,
	}
	self.parameters = parametersFromPath(self.path)
	self.handlers = make(map[string]http.HandlerFunc, 5)

	// Return success
	return self
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Handler returns the http.HandlerFunc for this path
func (p *Path) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if handler, ok := p.handlers[r.Method]; ok {
			handler(w, r)
			return
		}
		_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
	}
}

// Spec returns the openapi.PathItem for this path
func (p *Path) Spec() (string, *openapi.PathItem) {
	return p.path, types.Ptr(p.spec)
}

// Register registers an http.HandlerFunc for the given method
func (p *Path) Register(method string, handler http.HandlerFunc, summary string) error {
	// Register the handler
	p.handlers[method] = handler

	// Register the method in the spec
	switch strings.ToUpper(method) {
	case http.MethodGet:
		p.spec.Get = &openapi.Operation{Summary: summary, Parameters: p.parameters}
	case http.MethodPost:
		p.spec.Post = &openapi.Operation{Summary: summary, Parameters: p.parameters}
	case http.MethodPut:
		p.spec.Put = &openapi.Operation{Summary: summary, Parameters: p.parameters}
	case http.MethodPatch:
		p.spec.Patch = &openapi.Operation{Summary: summary, Parameters: p.parameters}
	case http.MethodDelete:
		p.spec.Delete = &openapi.Operation{Summary: summary, Parameters: p.parameters}
	case http.MethodHead:
		p.spec.Head = &openapi.Operation{Summary: summary, Parameters: p.parameters}
	case http.MethodOptions:
		p.spec.Options = &openapi.Operation{Summary: summary, Parameters: p.parameters}
	case http.MethodTrace:
		p.spec.Trace = &openapi.Operation{Summary: summary, Parameters: p.parameters}
	default:
		return httpresponse.Err(http.StatusMethodNotAllowed).Withf("unsupported method %q", method)
	}

	// Success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func parametersFromPath(path string) []openapi.Parameter {
	path = strings.TrimSpace(types.NormalisePath(path))
	if path == "" || path == "/" {
		return nil
	}

	segments := strings.Split(path, "/")
	stringSchema := jsonschema.MustFor[string]()
	params := make([]openapi.Parameter, 0, len(segments))
	seen := make(map[string]struct{}, len(segments))

	for _, segment := range segments {
		if len(segment) < 3 || segment[0] != '{' || segment[len(segment)-1] != '}' {
			continue
		}
		name := strings.TrimSpace(segment[1 : len(segment)-1])
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		params = append(params, openapi.Parameter{
			Name:     name,
			In:       openapi.ParameterInPath,
			Required: true,
			Schema:   stringSchema,
		})
	}

	return params
}
