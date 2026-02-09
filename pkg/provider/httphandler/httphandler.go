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
// PUBLIC METHODS

// RegisterHandlers registers HTTP handlers for provider operations on the
// given router under the given prefix.
func RegisterHandlers(router *httprouter.Router, prefix string, middleware bool, manager *provider.Manager) {
	// GET /resource — list all resource types and their schemas
	// POST /resource — create a new resource
	router.RegisterFunc(types.JoinPath(prefix, "resource"), func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			resp, err := manager.ListResources(r.Context(), schema.ListResourcesRequest{})
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
	}, middleware, types.Ptr(openapi.PathItem{
		Get: &openapi.Operation{
			Description: types.Ptr("List all resource types and their schemas"),
		}, Post: &openapi.Operation{
			Description: types.Ptr("Create a new resource instance of the specified type"),
		},
	}))
}
