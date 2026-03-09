package httphandler

import (
	"net/http"

	// Packages
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	jsonschema "github.com/mutablelogic/go-server/pkg/jsonschema"
	openapi "github.com/mutablelogic/go-server/pkg/openapi/schema"
	provider "github.com/mutablelogic/go-server/pkg/provider"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// HANDLER FUNCTIONS

// ResourceListHandler returns an HTTP handler that lists resource types
// (GET) and creates new resource instances (POST).
func ResourceListHandler(manager *provider.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			var req schema.ListResourcesRequest
			if err := httprequest.Query(r.URL.Query(), &req); err != nil {
				_ = httpresponse.Error(w, err)
				return
			}
			resp, err := manager.ListResources(r.Context(), req)
			if err != nil {
				_ = httpresponse.Error(w, err)
				return
			}
			_ = httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), resp)
		case http.MethodPost:
			var req schema.CreateResourceInstanceRequest
			if err := httprequest.Read(r, &req); err != nil {
				_ = httpresponse.Error(w, err)
				return
			}
			resp, err := manager.CreateResourceInstance(r.Context(), req)
			if err != nil {
				_ = httpresponse.Error(w, err)
				return
			}
			_ = httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), resp)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	}
}

// ResourceListSpec returns the OpenAPI path-item for the resource list endpoint.
func ResourceListSpec() *openapi.PathItem {
	createSchema, _ := jsonschema.For[schema.CreateResourceInstanceRequest]()
	typeSchema, _ := jsonschema.For[string]()
	listRespSchema, _ := jsonschema.For[schema.ListResourcesResponse]()
	createRespSchema, _ := jsonschema.For[schema.CreateResourceInstanceResponse]()
	return types.Ptr(openapi.PathItem{
		Get: &openapi.Operation{
			Tags:        []string{"Resources"},
			Summary:     "List resources",
			Description: "Returns resource types, or resource instances when the ?type= query parameter is set.",
			Parameters: []openapi.Parameter{
				{
					Name:        "type",
					In:          openapi.ParameterInQuery,
					Description: "Filter by resource type name (e.g. \"httpserver\")",
					Schema:      typeSchema,
				},
			},
			Responses: map[string]openapi.Response{
				"200":     {Description: "OK", Content: map[string]openapi.MediaType{types.ContentTypeJSON: {Schema: listRespSchema}}},
				"default": openapi.ErrorResponse("Error"),
			},
		},
		Post: &openapi.Operation{
			Tags:        []string{"Resources"},
			Summary:     "Create resource instance",
			Description: "Creates a new instance of the specified resource type.",
			RequestBody: &openapi.RequestBody{
				Required: true,
				Content: map[string]openapi.MediaType{
					types.ContentTypeJSON: {Schema: createSchema},
				},
			},
			Responses: map[string]openapi.Response{
				"201":     {Description: "Created", Content: map[string]openapi.MediaType{types.ContentTypeJSON: {Schema: createRespSchema}}},
				"default": openapi.ErrorResponse("Error"),
			},
		},
	})
}

// ResourceInstanceHandler returns an HTTP handler for get (GET), plan/apply
// (PATCH), and destroy (DELETE) of a single resource instance.
func ResourceInstanceHandler(manager *provider.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			_ = httpresponse.Error(w, httpresponse.ErrBadRequest.With("missing resource id"))
			return
		}
		switch r.Method {
		case http.MethodGet:
			resp, err := manager.GetResourceInstance(r.Context(), id)
			if err != nil {
				_ = httpresponse.Error(w, err)
				return
			}
			_ = httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), resp)
		case http.MethodPatch:
			var req schema.UpdateResourceInstanceRequest
			if err := httprequest.Read(r, &req); err != nil {
				_ = httpresponse.Error(w, err)
				return
			}
			resp, err := manager.UpdateResourceInstance(r.Context(), id, req)
			if err != nil {
				_ = httpresponse.Error(w, err)
				return
			}
			_ = httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), resp)
		case http.MethodDelete:
			resp, err := manager.DestroyResourceInstance(r.Context(), schema.DestroyResourceInstanceRequest{
				Name:    id,
				Cascade: r.URL.Query().Get("cascade") == "true",
			})
			if err != nil {
				_ = httpresponse.Error(w, err)
				return
			}
			_ = httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), resp)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	}
}

// ResourceInstanceSpec returns the OpenAPI path-item for the single-instance endpoint.
func ResourceInstanceSpec() *openapi.PathItem {
	updateSchema, _ := jsonschema.For[schema.UpdateResourceInstanceRequest]()
	idSchema, _ := jsonschema.For[string]()
	boolSchema, _ := jsonschema.For[bool]()
	getRespSchema, _ := jsonschema.For[schema.GetResourceInstanceResponse]()
	updateRespSchema, _ := jsonschema.For[schema.UpdateResourceInstanceResponse]()
	destroyRespSchema, _ := jsonschema.For[schema.DestroyResourceInstanceResponse]()
	idParam := openapi.Parameter{
		Name:        "id",
		In:          openapi.ParameterInPath,
		Description: "Resource instance ID",
		Required:    true,
		Schema:      idSchema,
	}
	return types.Ptr(openapi.PathItem{
		Get: &openapi.Operation{
			Tags:        []string{"Resources"},
			Summary:     "Get resource instance",
			Description: "Returns the metadata and current state of a single resource instance by ID.",
			Parameters:  []openapi.Parameter{idParam},
			Responses: map[string]openapi.Response{
				"200":     {Description: "OK", Content: map[string]openapi.MediaType{types.ContentTypeJSON: {Schema: getRespSchema}}},
				"default": openapi.ErrorResponse("Error"),
			},
		},
		Patch: &openapi.Operation{
			Tags:        []string{"Resources"},
			Summary:     "Plan or apply changes",
			Description: "Computes a plan for the desired attribute values. When apply is true, the plan is executed immediately.",
			Parameters:  []openapi.Parameter{idParam},
			RequestBody: &openapi.RequestBody{
				Required: true,
				Content: map[string]openapi.MediaType{
					types.ContentTypeJSON: {Schema: updateSchema},
				},
			},
			Responses: map[string]openapi.Response{
				"200":     {Description: "OK", Content: map[string]openapi.MediaType{types.ContentTypeJSON: {Schema: updateRespSchema}}},
				"default": openapi.ErrorResponse("Error"),
			},
		},
		Delete: &openapi.Operation{
			Tags:        []string{"Resources"},
			Summary:     "Destroy resource instance",
			Description: "Tears down the resource instance and releases its resources. Set cascade=true to also destroy dependents.",
			Parameters: []openapi.Parameter{
				idParam,
				{
					Name:        "cascade",
					In:          openapi.ParameterInQuery,
					Description: "Also destroy dependent instances in topological order",
					Schema:      boolSchema,
				},
			},
			Responses: map[string]openapi.Response{
				"200":     {Description: "OK", Content: map[string]openapi.MediaType{types.ContentTypeJSON: {Schema: destroyRespSchema}}},
				"default": openapi.ErrorResponse("Error"),
			},
		},
	})
}
