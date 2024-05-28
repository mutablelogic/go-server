package nginx

import (
	"context"
	"net/http"
	"regexp"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	router "github.com/mutablelogic/go-server/pkg/handler/router"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type responseHealth struct {
	Version string `json:"version"`
	Uptime  uint64 `json:"uptime"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	reRoot       = regexp.MustCompile(`^/?$`)
	reAction     = regexp.MustCompile(`^/(test|reload|reopen)/?$`)
	reListConfig = regexp.MustCompile(`^/config/?$`)
	reConfig     = regexp.MustCompile(`^/config/(.*)$`)
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - ENDPOINTS

// Add endpoints to the router
func (service *nginx) AddEndpoints(ctx context.Context, router server.Router) {
	// Path: /
	// Methods: GET
	// Scopes: read // TODO: Add scopes
	// Description: Get nginx status (version, uptime, available and enabled configurations)
	router.AddHandlerFuncRe(ctx, reRoot, service.GetHealth, http.MethodGet)

	// Path: /(test|reload|reopen)
	// Methods: PUT
	// Scopes: write // TODO: Add scopes
	// Description: Test, reload and reopen nginx configuration
	router.AddHandlerFuncRe(ctx, reAction, service.PutAction, http.MethodPut)

	// Path: /config
	// Methods: GET
	// Scopes: read // TODO: Add scopes
	// Description: Read the current set of configurations
	router.AddHandlerFuncRe(ctx, reListConfig, service.ListConfig, http.MethodGet)

	// Path: /config/{id}
	// Methods: GET
	// Scopes: read // TODO: Add scopes
	// Description: Read a configuration
	router.AddHandlerFuncRe(ctx, reConfig, service.ReadConfig, http.MethodGet)

	// Path: /config/{id}
	// Methods: DELETE, POST, PATCH
	// Scopes: write // TODO: Add scopes
	// Description: Modify a configuration
	router.AddHandlerFuncRe(ctx, reConfig, service.WriteConfig, http.MethodDelete, http.MethodPost, http.MethodPatch)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (service *nginx) GetHealth(w http.ResponseWriter, r *http.Request) {
	httpresponse.JSON(w, responseHealth{
		Version: string(service.Version()),
		Uptime:  uint64(time.Since(service.run.Start).Seconds()),
	}, http.StatusOK, 2)
}

func (service *nginx) PutAction(w http.ResponseWriter, r *http.Request) {
	urlParameters := router.Params(r.Context())
	switch urlParameters[0] {
	case "test":
		if err := service.Test(); err != nil {
			httpresponse.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
	case "reload":
		if err := service.Reload(); err != nil {
			httpresponse.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
	case "reopen":
		if err := service.Reopen(); err != nil {
			httpresponse.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
	default:
		httpresponse.Error(w, http.StatusNotFound)
		return
	}

	// Serve OK response with no body
	httpresponse.Empty(w, http.StatusOK)
}

func (service *nginx) ListConfig(w http.ResponseWriter, r *http.Request) {
	httpresponse.JSON(w, service.folders.Templates(), http.StatusOK, 2)
}

func (service *nginx) ReadConfig(w http.ResponseWriter, r *http.Request) {
	urlParameters := router.Params(r.Context())
	templ := service.folders.Template(urlParameters[0])
	if templ == nil {
		httpresponse.Error(w, http.StatusNotFound, urlParameters[0])
		return
	}
	httpresponse.JSON(w, templ, http.StatusOK, 2)
}

func (service *nginx) WriteConfig(w http.ResponseWriter, r *http.Request) {
	urlParameters := router.Params(r.Context())
	templ := service.folders.Template(urlParameters[0])
	if templ == nil {
		httpresponse.Error(w, http.StatusNotFound, urlParameters[0])
		return
	}
	httpresponse.Error(w, http.StatusNotImplemented)
}
