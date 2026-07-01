package httprequest

import (
	"net/http"
	"slices"
	"strconv"
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
	// Add one or more tags to the path item.
	// Tags are used to group operations in the OpenAPI document.
	Tag(...string) PathItem

	// Get registers a GET handler with the given summary and operation options.
	Get(handler http.HandlerFunc, fn func(PathOperation)) PathItem

	// Put registers a PUT handler with the given summary and operation options.
	Put(handler http.HandlerFunc, fn func(PathOperation)) PathItem

	// Post registers a POST handler with the given summary and operation options.
	Post(handler http.HandlerFunc, fn func(PathOperation)) PathItem

	// Delete registers a DELETE handler with the given summary and operation options.
	Delete(handler http.HandlerFunc, fn func(PathOperation)) PathItem

	// Patch registers a PATCH handler with the given summary and operation options.
	Patch(handler http.HandlerFunc, fn func(PathOperation)) PathItem

	// Options registers an OPTIONS handler with the given summary and operation options.
	Options(handler http.HandlerFunc, fn func(PathOperation)) PathItem

	// Head registers a HEAD handler with the given summary and operation options.
	Head(handler http.HandlerFunc, fn func(PathOperation)) PathItem

	// Trace registers a TRACE handler with the given summary and operation options.
	Trace(handler http.HandlerFunc, fn func(PathOperation)) PathItem

	// Handler returns the http.HandlerFunc that handles the PathItem
	Handler() http.HandlerFunc

	// Spec returns the openapi.PathItem for a path with optional path parameters
	Spec(string, *jsonschema.Schema) *openapi.PathItem

	// WrapHandler wraps the handler for a specific HTTP method with middleware.
	// The method should be an uppercase HTTP method name (GET, POST, etc.).
	// If no handler is registered for the method, the call is a no-op.
	WrapHandler(method string, fn func(http.HandlerFunc) http.HandlerFunc)
}

type PathOperation interface {
	// Add one or more tags to the operation.
	Tags(...string) PathOperation

	// Set the summary for the operation.
	Summary(string) PathOperation

	// Set the description for the operation.
	Description(string) PathOperation

	// Set the query parameters for the operation from a JSON schema.
	Query(*jsonschema.Schema) PathOperation

	// Mark the operation as deprecated.
	Deprecated() PathOperation

	// Add a JSON response for the operation with the given status code and schema.
	// An optional description can be provided; if not, the default HTTP status text will be used.
	JSONResponse(status int, schema *jsonschema.Schema, description ...string) PathOperation

	// Add an error response for the operation with the given status code.
	// An optional description can be provided; if not, the default HTTP status text will be used.
	ErrorResponse(status int, description ...string) PathOperation

	// Return a response for the operation with the given status code and content type.
	// An optional description can be provided; if not, the default HTTP status text will be used.
	Response(status int, contentType string, description ...string) PathOperation
}

///////////////////////////////////////////////////////////////////////////////
// TYPES

type pathitem struct {
	spec     openapi.PathItem
	tags     []string
	handlers map[string]http.HandlerFunc
}

type pathoperation struct {
	spec *openapi.Operation
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
// PUBLIC METHODS - PATH ITEM

// Add one or more tags to the path item.
// Tags are used to group operations in the OpenAPI document.
func (p *pathitem) Tag(tags ...string) PathItem {
	p.tags = append(p.tags, tags...)
	return p
}

// Return a new PathOperation for the path item.
func (p *pathitem) Get(handler http.HandlerFunc, fn func(PathOperation)) PathItem {
	p.handlers[http.MethodGet] = handler
	p.spec.Get = types.Ptr(openapi_op.Operation(http.MethodGet, openapi_op.WithTags(p.tags...)))
	if fn != nil {
		fn(&pathoperation{spec: p.spec.Get})
	}
	return p
}

// Return a new PathOperation for the path item.
func (p *pathitem) Put(handler http.HandlerFunc, fn func(PathOperation)) PathItem {
	p.handlers[http.MethodPut] = handler
	p.spec.Put = types.Ptr(openapi_op.Operation(http.MethodPut, openapi_op.WithTags(p.tags...)))
	if fn != nil {
		fn(&pathoperation{spec: p.spec.Put})
	}
	return p
}

// Return a new PathOperation for the path item.
func (p *pathitem) Post(handler http.HandlerFunc, fn func(PathOperation)) PathItem {
	p.handlers[http.MethodPost] = handler
	p.spec.Post = types.Ptr(openapi_op.Operation(http.MethodPost, openapi_op.WithTags(p.tags...)))
	if fn != nil {
		fn(&pathoperation{spec: p.spec.Post})
	}
	return p
}

// Return a new PathOperation for the path item.
func (p *pathitem) Delete(handler http.HandlerFunc, fn func(PathOperation)) PathItem {
	p.handlers[http.MethodDelete] = handler
	p.spec.Delete = types.Ptr(openapi_op.Operation(http.MethodDelete, openapi_op.WithTags(p.tags...)))
	if fn != nil {
		fn(&pathoperation{spec: p.spec.Delete})
	}
	return p
}

// Return a new PathOperation for the path item.
func (p *pathitem) Patch(handler http.HandlerFunc, fn func(PathOperation)) PathItem {
	p.handlers[http.MethodPatch] = handler
	p.spec.Patch = types.Ptr(openapi_op.Operation(http.MethodPatch, openapi_op.WithTags(p.tags...)))
	if fn != nil {
		fn(&pathoperation{spec: p.spec.Patch})
	}
	return p
}

// Return a new PathOperation for the path item.
func (p *pathitem) Options(handler http.HandlerFunc, fn func(PathOperation)) PathItem {
	p.handlers[http.MethodOptions] = handler
	p.spec.Options = types.Ptr(openapi_op.Operation(http.MethodOptions, openapi_op.WithTags(p.tags...)))
	if fn != nil {
		fn(&pathoperation{spec: p.spec.Options})
	}
	return p
}

// Return a new PathOperation for the path item.
func (p *pathitem) Head(handler http.HandlerFunc, fn func(PathOperation)) PathItem {
	p.handlers[http.MethodHead] = handler
	p.spec.Head = types.Ptr(openapi_op.Operation(http.MethodHead, openapi_op.WithTags(p.tags...)))
	if fn != nil {
		fn(&pathoperation{spec: p.spec.Head})
	}
	return p
}

// Return a new PathOperation for the path item.
func (p *pathitem) Trace(handler http.HandlerFunc, fn func(PathOperation)) PathItem {
	p.handlers[http.MethodTrace] = handler
	p.spec.Trace = types.Ptr(openapi_op.Operation(http.MethodTrace, openapi_op.WithTags(p.tags...)))
	if fn != nil {
		fn(&pathoperation{spec: p.spec.Trace})
	}
	return p
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

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PATH OPERATION

func (p *pathoperation) Tags(tags ...string) PathOperation {
	p.spec.Tags = append(p.spec.Tags, tags...)
	return p
}

func (p *pathoperation) Summary(summary string) PathOperation {
	p.spec.Summary = summary
	return p
}

func (p *pathoperation) Description(description string) PathOperation {
	p.spec.Description = description
	return p
}

func (p *pathoperation) Query(q *jsonschema.Schema) PathOperation {
	p.spec.Parameters = append(p.spec.Parameters, parametersFromQuery(q)...)
	return p
}

func (p *pathoperation) Deprecated() PathOperation {
	p.spec.Deprecated = true
	return p
}

func (p *pathoperation) JSONResponse(status int, schema *jsonschema.Schema, description ...string) PathOperation {
	if p.spec.Responses == nil {
		p.spec.Responses = make(map[string]openapi.Response)
	}
	statusNum := strconv.FormatInt(int64(status), 10)
	descriptionText := strings.Join(description, " ")
	if descriptionText == "" {
		descriptionText = http.StatusText(status)
	}
	p.spec.Responses[statusNum] = openapi.Response{
		Description: descriptionText,
		Content: map[string]openapi.MediaType{
			types.ContentTypeJSON: {
				Schema: schema,
			},
		},
	}
	return p
}

func (p *pathoperation) ErrorResponse(status int, description ...string) PathOperation {
	return p.JSONResponse(status, jsonschema.MustFor[httpresponse.ErrResponse](), description...)
}

func (p *pathoperation) Response(status int, contentType string, description ...string) PathOperation {
	if p.spec.Responses == nil {
		p.spec.Responses = make(map[string]openapi.Response)
	}
	statusNum := strconv.FormatInt(int64(status), 10)
	descriptionText := strings.Join(description, " ")
	if descriptionText == "" {
		descriptionText = http.StatusText(status)
	}
	p.spec.Responses[statusNum] = openapi.Response{
		Description: descriptionText,
		Content: map[string]openapi.MediaType{
			contentType: {},
		},
	}
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

func parametersFromQuery(q *jsonschema.Schema) []openapi.Parameter {
	if q == nil {
		return nil
	}
	names := make([]string, 0, len(q.Properties))
	for name := range q.Properties {
		names = append(names, name)
	}
	slices.Sort(names)
	params := make([]openapi.Parameter, 0, len(names))
	for _, name := range names {
		params = append(params, openapi.Parameter{
			Name:     name,
			In:       openapi.ParameterInQuery,
			Required: slices.Contains(q.Required, name),
			Schema:   q.Property(name),
		})
	}
	return params
}
