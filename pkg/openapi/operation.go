package openapi

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	// Packages
	upstreamjsonschema "github.com/google/jsonschema-go/jsonschema"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	jsonschema "github.com/mutablelogic/go-server/pkg/jsonschema"
	schema "github.com/mutablelogic/go-server/pkg/openapi/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// OperationOpt applies a configuration change to an OpenAPI operation.
type OperationOpt func(*opopt)

type opopt struct {
	id          string
	description string
	deprecated  bool
	tags        []string
	query       []schema.Parameter
	request     schema.RequestBody
	responses   map[string]schema.Response // keyed by status code or "default"
	security    []schema.SecurityRequirement
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var stringSchema = jsonschema.MustFor[string]()

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Operation builds an OpenAPI operation from a summary and option list.
func Operation(summary string, opts ...OperationOpt) schema.Operation {
	operation := schema.Operation{
		Summary: summary,
	}

	// Apply options
	o := opopt{
		request: schema.RequestBody{
			Content: make(map[string]schema.MediaType),
		},
		responses: make(map[string]schema.Response),
	}
	for _, opt := range opts {
		opt(&o)
	}

	// Add tags, query parameters, requests and responses
	operation.OperationId = o.id
	operation.Description = o.description
	operation.Deprecated = o.deprecated
	operation.Tags = o.tags
	operation.Parameters = o.query
	if len(o.request.Content) > 0 {
		operation.RequestBody = types.Ptr(o.request)
	}
	if len(o.responses) > 0 {
		operation.Responses = o.responses
	}
	if len(o.security) > 0 {
		operation.Security = o.security
	}

	// Return the operation
	return operation
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - OPTIONS

// WithDescription sets the operation description.
func WithDescription(description string) OperationOpt {
	return func(opt *opopt) {
		opt.description = description
	}
}

// WithID sets the operation identifier.
func WithID(id string) OperationOpt {
	return func(opt *opopt) {
		opt.id = id
	}
}

// WithDeprecated marks the operation as deprecated.
func WithDeprecated() OperationOpt {
	return func(opt *opopt) {
		opt.deprecated = true
	}
}

// WithTags appends tags to the operation.
func WithTags(tags ...string) OperationOpt {
	return func(opt *opopt) {
		opt.tags = append(opt.tags, tags...)
	}
}

// WithQuery adds query parameters inferred from a JSON schema.
func WithQuery(q *jsonschema.Schema) OperationOpt {
	return func(opt *opopt) {
		opt.query = append(opt.query, parametersFromQuery(q)...)
	}
}

// WithRequest sets the request body schema for a specific content type.
//
// When more than one schema is provided, they are wrapped in a `oneOf` schema.
func WithRequest(contentType string, schemas ...*jsonschema.Schema) OperationOpt {
	return func(opt *opopt) {
		if r := requestSchema("", schemas...); r != nil {
			opt.request.Content[contentType] = schema.MediaType{
				Schema: r,
			}
		}
	}
}

// WithNamedRequest sets the request body schema for a specific content type and
// applies a title to the generated wrapper schema.
//
// When more than one schema is provided, they are wrapped in a titled `oneOf`
// schema.
func WithNamedRequest(contentType, title string, schemas ...*jsonschema.Schema) OperationOpt {
	return func(opt *opopt) {
		if r := requestSchema(title, schemas...); r != nil {
			opt.request.Content[contentType] = schema.MediaType{
				Schema: r,
			}
		}
	}
}

// WithJSONRequest sets the JSON request body schema.
func WithJSONRequest(schemas ...*jsonschema.Schema) OperationOpt {
	return WithRequest(types.ContentTypeJSON, schemas...)
}

// WithNamedJSONRequest sets a titled JSON request body schema.
func WithNamedJSONRequest(title string, schemas ...*jsonschema.Schema) OperationOpt {
	return WithNamedRequest(types.ContentTypeJSON, title, schemas...)
}

// WithFormRequest sets the form-encoded request body schema.
func WithFormRequest(schemas ...*jsonschema.Schema) OperationOpt {
	return WithRequest(types.ContentTypeForm, schemas...)
}

// WithNamedFormRequest sets a titled form-encoded request body schema.
func WithNamedFormRequest(title string, schemas ...*jsonschema.Schema) OperationOpt {
	return WithNamedRequest(types.ContentTypeForm, title, schemas...)
}

// WithMultipartRequest sets the multipart form request body schema.
func WithMultipartRequest(schemas ...*jsonschema.Schema) OperationOpt {
	return WithRequest(types.ContentTypeFormData, schemas...)
}

// WithNamedMultipartRequest sets a titled multipart form request body schema.
func WithNamedMultipartRequest(title string, schemas ...*jsonschema.Schema) OperationOpt {
	return WithNamedRequest(types.ContentTypeFormData, title, schemas...)
}

// WithJSONResponse sets the JSON response schema for a status code.
func WithJSONResponse(status int, r *jsonschema.Schema) OperationOpt {
	return WithResponse(status, types.ContentTypeJSON, r)
}

// WithTextResponse sets a plain-text response for a status code.
func WithTextResponse(status int, description string) OperationOpt {
	return WithResponse(status, types.ContentTypeTextPlain, stringSchema, description)
}

// WithTextStreamResponse sets a text/event-stream style response for a status code.
func WithTextStreamResponse(status int, description ...string) OperationOpt {
	return WithResponse(status, types.ContentTypeTextStream, stringSchema, description...)
}

// WithNoContentResponse sets a response with no body for a status code.
func WithNoContentResponse(status int, description ...string) OperationOpt {
	statusString := func(status int) string {
		if status < 100 || status > 599 {
			return "default"
		}
		return fmt.Sprintf("%d", status)
	}
	descriptionString := func() string {
		if len(description) > 0 {
			return strings.Join(description, " ")
		}
		return http.StatusText(status)
	}
	return func(opt *opopt) {
		opt.responses[statusString(status)] = schema.Response{
			Description: descriptionString(),
		}
	}
}

// WithResponse sets the response schema for a status code and content type.
func WithResponse(status int, contentType string, r *jsonschema.Schema, description ...string) OperationOpt {
	statusString := func(status int) string {
		if status < 100 || status > 599 {
			return "default"
		}
		return fmt.Sprintf("%d", status)
	}
	descriptionString := func() string {
		if len(description) > 0 {
			return strings.Join(description, " ")
		}
		return http.StatusText(status)
	}
	return func(opt *opopt) {
		if r != nil {
			key := statusString(status)
			response, exists := opt.responses[key]
			if !exists {
				response = schema.Response{
					Description: descriptionString(),
				}
			}
			if response.Content == nil {
				response.Content = make(map[string]schema.MediaType)
			}
			response.Content[contentType] = schema.MediaType{
				Schema: r,
			}
			opt.responses[key] = response
		}
	}
}

// WithSecurity adds a security requirement to the operation.
func WithSecurity(schemeName string, enabled bool, scopes ...string) OperationOpt {
	return func(opt *opopt) {
		if enabled {
			opt.security = append(opt.security, schema.SecurityRequirement{
				schemeName: scopes,
			})
		}
	}
}

// WithErrorResponse sets a JSON error response using httpresponse.ErrResponse.
func WithErrorResponse(status int, description ...string) OperationOpt {
	statusString := func() string {
		if status < 100 || status > 599 {
			return "default"
		}
		return fmt.Sprintf("%d", status)
	}
	descriptionString := func() string {
		if len(description) == 0 {
			return http.StatusText(status)
		}
		return strings.Join(description, " ")
	}
	return func(opt *opopt) {
		key := statusString()
		opt.responses[key] = schema.Response{
			Description: descriptionString(),
			Content: map[string]schema.MediaType{
				types.ContentTypeJSON: {
					Schema: jsonschema.MustFor[httpresponse.ErrResponse](),
				},
			},
		}
	}
}

// NamedSchema clones a schema and applies a title to the clone.
func NamedSchema(title string, schema *jsonschema.Schema) *jsonschema.Schema {
	if schema == nil {
		return nil
	}
	clone := *schema
	clone.Title = title
	return &clone
}

func requestSchema(title string, schemas ...*jsonschema.Schema) *jsonschema.Schema {
	filtered := make([]*jsonschema.Schema, 0, len(schemas))
	for _, schema := range schemas {
		if schema != nil {
			filtered = append(filtered, schema)
		}
	}
	switch len(filtered) {
	case 0:
		return nil
	case 1:
		return filtered[0]
	default:
		oneOf := make([]*upstreamjsonschema.Schema, 0, len(filtered))
		for _, schema := range filtered {
			oneOf = append(oneOf, &schema.Schema)
		}
		return &jsonschema.Schema{
			Schema: upstreamjsonschema.Schema{
				Title: title,
				OneOf: oneOf,
			},
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func parametersFromQuery(q *jsonschema.Schema) []schema.Parameter {
	if q == nil {
		return nil
	}
	names := make([]string, 0, len(q.Properties))
	for name := range q.Properties {
		names = append(names, name)
	}
	slices.Sort(names)
	params := make([]schema.Parameter, 0, len(names))
	for _, name := range names {
		params = append(params, schema.Parameter{
			Name:     name,
			In:       schema.ParameterInQuery,
			Required: slices.Contains(q.Required, name),
			Schema:   q.Property(name),
		})
	}
	return params
}
