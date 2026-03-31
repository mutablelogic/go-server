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

// Handler returns the http.HandlerFunc for this path
func (p *pathitem) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if handler, ok := p.handlers[strings.ToUpper(strings.TrimSpace(r.Method))]; ok {
			handler(w, r)
			return
		}
		_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
	}
}

// Spec returns the openapi.PathItem for this path
func (p *pathitem) Spec(path string, params *jsonschema.Schema) *openapi.PathItem {
	p.spec.Parameters = parametersFromPath(path, params)
	return types.Ptr(p.spec)
}

// Register Get handler
func (p *pathitem) Get(handler http.HandlerFunc, summary string, opts ...openapi_op.OperationOpt) *pathitem {
	p.handlers[http.MethodGet] = handler
	p.spec.Get = types.Ptr(openapi_op.Operation(summary, append([]openapi_op.OperationOpt{openapi_op.WithTags(p.tags...)}, opts...)...))
	return p
}

// Register Put handler
func (p *pathitem) Put(handler http.HandlerFunc, summary string, opts ...openapi_op.OperationOpt) *pathitem {
	p.handlers[http.MethodPut] = handler
	p.spec.Put = types.Ptr(openapi_op.Operation(summary, append([]openapi_op.OperationOpt{openapi_op.WithTags(p.tags...)}, opts...)...))
	return p
}

// Register Post handler
func (p *pathitem) Post(handler http.HandlerFunc, summary string, opts ...openapi_op.OperationOpt) *pathitem {
	p.handlers[http.MethodPost] = handler
	p.spec.Post = types.Ptr(openapi_op.Operation(summary, append([]openapi_op.OperationOpt{openapi_op.WithTags(p.tags...)}, opts...)...))
	return p
}

// Register Delete handler
func (p *pathitem) Delete(handler http.HandlerFunc, summary string, opts ...openapi_op.OperationOpt) *pathitem {
	p.handlers[http.MethodDelete] = handler
	p.spec.Delete = types.Ptr(openapi_op.Operation(summary, append([]openapi_op.OperationOpt{openapi_op.WithTags(p.tags...)}, opts...)...))
	return p
}

// Register Patch handler
func (p *pathitem) Patch(handler http.HandlerFunc, summary string, opts ...openapi_op.OperationOpt) *pathitem {
	p.handlers[http.MethodPatch] = handler
	p.spec.Patch = types.Ptr(openapi_op.Operation(summary, append([]openapi_op.OperationOpt{openapi_op.WithTags(p.tags...)}, opts...)...))
	return p
}

// Register Options handler
func (p *pathitem) Options(handler http.HandlerFunc, summary string, opts ...openapi_op.OperationOpt) *pathitem {
	p.handlers[http.MethodOptions] = handler
	p.spec.Options = types.Ptr(openapi_op.Operation(summary, append([]openapi_op.OperationOpt{openapi_op.WithTags(p.tags...)}, opts...)...))
	return p
}

// Register Head handler
func (p *pathitem) Head(handler http.HandlerFunc, summary string, opts ...openapi_op.OperationOpt) *pathitem {
	p.handlers[http.MethodHead] = handler
	p.spec.Head = types.Ptr(openapi_op.Operation(summary, append([]openapi_op.OperationOpt{openapi_op.WithTags(p.tags...)}, opts...)...))
	return p
}

// Register Trace handler
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
