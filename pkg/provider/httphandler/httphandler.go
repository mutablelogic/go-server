package httphandler

import (
	"net/http"

	// Packages
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
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
			_ = httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), resp)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	}
}

// ResourceListSpec returns the OpenAPI path-item for the resource list endpoint.
func ResourceListSpec() *openapi.PathItem {
	return types.Ptr(openapi.PathItem{
		Get: &openapi.Operation{
			Description: types.Ptr("List resource types, or instances when ?resource= is set"),
		},
		Post: &openapi.Operation{
			Description: types.Ptr("Create a new resource instance of the specified type"),
		},
	})
}

// ResourceInstanceHandler returns an HTTP handler for get (GET), plan/apply
// (PATCH), and destroy (DELETE) of a single resource instance.
func ResourceInstanceHandler(manager *provider.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			resp, err := manager.GetResourceInstance(r.Context(), r.PathValue("id"))
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
			resp, err := manager.UpdateResourceInstance(r.Context(), r.PathValue("id"), req)
			if err != nil {
				_ = httpresponse.Error(w, err)
				return
			}
			_ = httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), resp)
		case http.MethodDelete:
			resp, err := manager.DestroyResourceInstance(r.Context(), schema.DestroyResourceInstanceRequest{
				Name:    r.PathValue("id"),
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
	return types.Ptr(openapi.PathItem{
		Get: &openapi.Operation{
			Description: types.Ptr("Get a resource instance by ID"),
		},
		Patch: &openapi.Operation{
			Description: types.Ptr("Plan or apply changes to a resource instance"),
		},
		Delete: &openapi.Operation{
			Description: types.Ptr("Destroy a resource instance"),
		},
	})
}

///////////////////////////////////////////////////////////////////////////////
// LEGACY

// RegisterHandlers registers HTTP handlers for provider operations on the
// given router under the given prefix.
func RegisterHandlers(router *httprouter.Router, prefix string, middleware bool, manager *provider.Manager) {
	router.RegisterFunc(
		types.JoinPath(prefix, "resource"),
		ResourceListHandler(manager),
		middleware,
		ResourceListSpec(),
	)
	router.RegisterFunc(
		types.JoinPath(prefix, "resource/{id}"),
		ResourceInstanceHandler(manager),
		middleware,
		ResourceInstanceSpec(),
	)
}
