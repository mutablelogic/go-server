package httprequest

import (
	"net/http"
	"strings"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	jsonschema "github.com/mutablelogic/go-server/pkg/jsonschema"
	openapi_op "github.com/mutablelogic/go-server/pkg/openapi"
	openapi "github.com/mutablelogic/go-server/pkg/openapi/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

type PathItem interface {
	// Handler returns the http.HandlerFunc that handles the PathItem
	Handler() http.HandlerFunc

	// Spec returns the openapi.PathItem for a path with optional path parameters
	Spec(string, *jsonschema.Schema) *openapi.PathItem

	// WrapHandler wraps the handler for a specific HTTP method with middleware.
	// The method should be an uppercase HTTP method name (GET, POST, etc.).
	// If no handler is registered for the method, the call is a no-op.
	WrapHandler(method string, fn func(http.HandlerFunc) http.HandlerFunc)
}

///////////////////////////////////////////////////////////////////////////////
// TYPES

type pathitem struct {
	spec     openapi.PathItem
	tags     []string
	handlers map[string]http.HandlerFunc
}

var _ PathItem = (*pathitem)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewPathItem creates a new path item with the given summary, description and tags.
func NewPathItem(summary, description string, tags ...string) *pathitem {
	self := new(pathitem)

	// Set the spec, parameters and handler
	self.spec = openapi.PathItem{
		Summary:     summary,
		Description: description,
	}
	self.handlers = make(map[string]http.HandlerFunc, 5)
	self.tags = tags

	// Return success
	return self
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Handler returns an http.HandlerFunc that dispatches to the registered method handler.
// Methods registered with a nil handler are treated as spec-only (they appear in the
// OpenAPI document but return 405 Method Not Allowed at runtime).
func (p *pathitem) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		method := strings.ToUpper(r.Method)
		if handler, ok := p.handlers[method]; ok && handler != nil {
			handler(w, r)
			return
		}
		_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
	}
}

// Spec returns the OpenAPI PathItem with path parameters resolved from the pattern.
func (p *pathitem) Spec(path string, params *jsonschema.Schema) *openapi.PathItem {
	p.spec.Parameters = parametersFromPath(path, params)
	return types.Ptr(p.spec)
}

// WrapHandler wraps the handler for a specific HTTP method with the given middleware function.
func (p *pathitem) WrapHandler(method string, fn func(http.HandlerFunc) http.HandlerFunc) {
	if h, ok := p.handlers[method]; ok && h != nil {
		p.handlers[method] = fn(h)
	}
}

// Get registers a GET handler with the given summary and operation options.
func (p *pathitem) Get(handler http.HandlerFunc, summary string, opts ...openapi_op.OperationOpt) *pathitem {
	p.handlers[http.MethodGet] = handler
	p.spec.Get = types.Ptr(openapi_op.Operation(summary, append([]openapi_op.OperationOpt{openapi_op.WithTags(p.tags...)}, opts...)...))
	return p
}

// Put registers a PUT handler with the given summary and operation options.
func (p *pathitem) Put(handler http.HandlerFunc, summary string, opts ...openapi_op.OperationOpt) *pathitem {
	p.handlers[http.MethodPut] = handler
	p.spec.Put = types.Ptr(openapi_op.Operation(summary, append([]openapi_op.OperationOpt{openapi_op.WithTags(p.tags...)}, opts...)...))
	return p
}

// Post registers a POST handler with the given summary and operation options.
func (p *pathitem) Post(handler http.HandlerFunc, summary string, opts ...openapi_op.OperationOpt) *pathitem {
	p.handlers[http.MethodPost] = handler
	p.spec.Post = types.Ptr(openapi_op.Operation(summary, append([]openapi_op.OperationOpt{openapi_op.WithTags(p.tags...)}, opts...)...))
	return p
}

// Delete registers a DELETE handler with the given summary and operation options.
func (p *pathitem) Delete(handler http.HandlerFunc, summary string, opts ...openapi_op.OperationOpt) *pathitem {
	p.handlers[http.MethodDelete] = handler
	p.spec.Delete = types.Ptr(openapi_op.Operation(summary, append([]openapi_op.OperationOpt{openapi_op.WithTags(p.tags...)}, opts...)...))
	return p
}

// Patch registers a PATCH handler with the given summary and operation options.
func (p *pathitem) Patch(handler http.HandlerFunc, summary string, opts ...openapi_op.OperationOpt) *pathitem {
	p.handlers[http.MethodPatch] = handler
	p.spec.Patch = types.Ptr(openapi_op.Operation(summary, append([]openapi_op.OperationOpt{openapi_op.WithTags(p.tags...)}, opts...)...))
	return p
}

// Options registers an OPTIONS handler with the given summary and operation options.
func (p *pathitem) Options(handler http.HandlerFunc, summary string, opts ...openapi_op.OperationOpt) *pathitem {
	p.handlers[http.MethodOptions] = handler
	p.spec.Options = types.Ptr(openapi_op.Operation(summary, append([]openapi_op.OperationOpt{openapi_op.WithTags(p.tags...)}, opts...)...))
	return p
}

// Head registers a HEAD handler with the given summary and operation options.
func (p *pathitem) Head(handler http.HandlerFunc, summary string, opts ...openapi_op.OperationOpt) *pathitem {
	p.handlers[http.MethodHead] = handler
	p.spec.Head = types.Ptr(openapi_op.Operation(summary, append([]openapi_op.OperationOpt{openapi_op.WithTags(p.tags...)}, opts...)...))
	return p
}

// Trace registers a TRACE handler with the given summary and operation options.
func (p *pathitem) Trace(handler http.HandlerFunc, summary string, opts ...openapi_op.OperationOpt) *pathitem {
	p.handlers[http.MethodTrace] = handler
	p.spec.Trace = types.Ptr(openapi_op.Operation(summary, append([]openapi_op.OperationOpt{openapi_op.WithTags(p.tags...)}, opts...)...))
	return p
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

var stringSchema = jsonschema.MustFor[string]()

func parametersFromPath(path string, schema *jsonschema.Schema) []openapi.Parameter {
	path = strings.TrimSpace(types.NormalisePath(path))
	if path == "" || path == "/" {
		return nil
	}

	segments := strings.Split(path, "/")
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

		if schema != nil {
			if prop := schema.Property(name); prop != nil {
				params[len(params)-1].Schema = prop
			}
		}

	}

	return params
}
