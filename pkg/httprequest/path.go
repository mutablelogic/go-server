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

type PathItem struct {
	spec     openapi.PathItem
	tags     []string
	handlers map[string]http.HandlerFunc
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewPathItem(summary, description string, tags ...string) *PathItem {
	self := new(PathItem)

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
func (p *PathItem) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if handler, ok := p.handlers[strings.ToUpper(strings.TrimSpace(r.Method))]; ok {
			handler(w, r)
			return
		}
		_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
	}
}

// Spec returns the openapi.PathItem for this path
func (p *PathItem) Spec(path string, params *jsonschema.Schema) *openapi.PathItem {
	p.spec.Parameters = parametersFromPath(path, params)
	return types.Ptr(p.spec)
}

// Register Get handler
func (p *PathItem) Get(handler http.HandlerFunc, summary string) *PathItem {
	p.handlers[http.MethodGet] = handler
	p.spec.Get = types.Ptr(openapi.Operation{Summary: summary, Tags: p.tags})
	return p
}

// Register Put handler
func (p *PathItem) Put(handler http.HandlerFunc, summary string) *PathItem {
	p.handlers[http.MethodPut] = handler
	p.spec.Put = types.Ptr(openapi.Operation{Summary: summary, Tags: p.tags})
	return p
}

// Register Post handler
func (p *PathItem) Post(handler http.HandlerFunc, summary string) *PathItem {
	p.handlers[http.MethodPost] = handler
	p.spec.Post = types.Ptr(openapi.Operation{Summary: summary, Tags: p.tags})
	return p
}

// Register Delete handler
func (p *PathItem) Delete(handler http.HandlerFunc, summary string) *PathItem {
	p.handlers[http.MethodDelete] = handler
	p.spec.Delete = types.Ptr(openapi.Operation{Summary: summary, Tags: p.tags})
	return p
}

// Register Patch handler
func (p *PathItem) Patch(handler http.HandlerFunc, summary string) *PathItem {
	p.handlers[http.MethodPatch] = handler
	p.spec.Patch = types.Ptr(openapi.Operation{Summary: summary, Tags: p.tags})
	return p
}

// Register Options handler
func (p *PathItem) Options(handler http.HandlerFunc, summary string) *PathItem {
	p.handlers[http.MethodOptions] = handler
	p.spec.Options = types.Ptr(openapi.Operation{Summary: summary, Tags: p.tags})
	return p
}

// Register Head handler
func (p *PathItem) Head(handler http.HandlerFunc, summary string) *PathItem {
	p.handlers[http.MethodHead] = handler
	p.spec.Head = types.Ptr(openapi.Operation{Summary: summary, Tags: p.tags})
	return p
}

// Register Trace handler
func (p *PathItem) Trace(handler http.HandlerFunc, summary string) *PathItem {
	p.handlers[http.MethodTrace] = handler
	p.spec.Trace = types.Ptr(openapi.Operation{Summary: summary, Tags: p.tags})
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
