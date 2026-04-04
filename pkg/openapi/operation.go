package openapi

import (
	// Packages
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/mutablelogic/go-server/pkg/jsonschema"
	"github.com/mutablelogic/go-server/pkg/openapi/schema"
	"github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

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

func WithDescription(description string) OperationOpt {
	return func(opt *opopt) {
		opt.description = description
	}
}

func WithID(id string) OperationOpt {
	return func(opt *opopt) {
		opt.id = id
	}
}

func WithDeprecated() OperationOpt {
	return func(opt *opopt) {
		opt.deprecated = true
	}
}

func WithTags(tags ...string) OperationOpt {
	return func(opt *opopt) {
		opt.tags = append(opt.tags, tags...)
	}
}

func WithQuery(q *jsonschema.Schema) OperationOpt {
	return func(opt *opopt) {
		opt.query = append(opt.query, parametersFromQuery(q)...)
	}
}

func WithRequest(contentType string, r *jsonschema.Schema) OperationOpt {
	return func(opt *opopt) {
		if r != nil {
			opt.request.Content[contentType] = schema.MediaType{
				Schema: r,
			}
		}
	}
}

func WithJSONRequest(r *jsonschema.Schema) OperationOpt {
	return WithRequest(types.ContentTypeJSON, r)
}

func WithFormRequest(r *jsonschema.Schema) OperationOpt {
	return WithRequest(types.ContentTypeForm, r)
}

func WithMultipartRequest(r *jsonschema.Schema) OperationOpt {
	return WithRequest(types.ContentTypeFormData, r)
}

func WithJSONResponse(status int, r *jsonschema.Schema) OperationOpt {
	return WithResponse(status, types.ContentTypeJSON, r)
}

func WithTextResponse(status int, description string) OperationOpt {
	return WithResponse(status, types.ContentTypeTextPlain, stringSchema, description)
}

func WithTextStreamResponse(status int, description ...string) OperationOpt {
	return WithResponse(status, types.ContentTypeTextStream, stringSchema, description...)
}

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

func WithSecurity(schemeName string, scopes ...string) OperationOpt {
	return func(opt *opopt) {
		opt.security = append(opt.security, schema.SecurityRequirement{
			schemeName: scopes,
		})
	}
}

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
