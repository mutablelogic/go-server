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
		if handler, ok := p.handlers[strings.ToUpper(strings.TrimSpace(r.Method))]; ok {
			handler(w, r)
			return
		}
		_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
	}
}

// Spec returns the openapi.PathItem for this path
func (p *Path) Spec() (string, *openapi.PathItem) {
	return p.path, clonePathItem(p.spec)
}

// Register registers an http.HandlerFunc for the given method
func (p *Path) Register(method string, handler http.HandlerFunc, summary string) error {
	method = strings.ToUpper(strings.TrimSpace(method))

	var operation **openapi.Operation

	// Register the method in the spec
	switch method {
	case http.MethodGet:
		operation = &p.spec.Get
	case http.MethodPost:
		operation = &p.spec.Post
	case http.MethodPut:
		operation = &p.spec.Put
	case http.MethodPatch:
		operation = &p.spec.Patch
	case http.MethodDelete:
		operation = &p.spec.Delete
	case http.MethodHead:
		operation = &p.spec.Head
	case http.MethodOptions:
		operation = &p.spec.Options
	case http.MethodTrace:
		operation = &p.spec.Trace
	default:
		return httpresponse.Err(http.StatusMethodNotAllowed).Withf("unsupported method %q", method)
	}

	// Register the handler and operation only after validation succeeds.
	p.handlers[method] = handler
	*operation = &openapi.Operation{Summary: summary, Parameters: p.parameters}

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

func clonePathItem(item openapi.PathItem) *openapi.PathItem {
	clone := item
	clone.Get = cloneOperation(item.Get)
	clone.Put = cloneOperation(item.Put)
	clone.Post = cloneOperation(item.Post)
	clone.Delete = cloneOperation(item.Delete)
	clone.Options = cloneOperation(item.Options)
	clone.Head = cloneOperation(item.Head)
	clone.Patch = cloneOperation(item.Patch)
	clone.Trace = cloneOperation(item.Trace)
	return types.Ptr(clone)
}

func cloneOperation(op *openapi.Operation) *openapi.Operation {
	if op == nil {
		return nil
	}
	clone := *op
	clone.Tags = append([]string(nil), op.Tags...)
	clone.Parameters = append([]openapi.Parameter(nil), op.Parameters...)
	return &clone
}
