package nginx

import (
	"context"
	"net/http"
	"regexp"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/httpresponse"
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
	reRoot   = regexp.MustCompile(`^/?$`)
	reAction = regexp.MustCompile(`^/(test|reload|reopen)/?$`)
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - ENDPOINTS

// Add endpoints to the router
func (service *nginx) AddEndpoints(ctx context.Context, router server.Router) {
	// Path: /
	// Methods: GET
	// Scopes: read
	// Description: Get nginx status (version, uptime, available and enabled configurations)
	router.AddHandlerFuncRe(ctx, reRoot, service.GetHealth, http.MethodGet)

	// Path: /(test|reload|reopen)
	// Methods: POST
	// Scopes: write
	// Description: Test, reload and reopen nginx configuration
	router.AddHandlerFuncRe(ctx, reAction, service.PostAction, http.MethodPost)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (service *nginx) GetHealth(w http.ResponseWriter, r *http.Request) {
	httpresponse.JSON(w, responseHealth{
		Version: string(service.Version()),
		Uptime:  uint64(time.Since(service.run.Start).Seconds()),
	}, http.StatusOK, 2)
}

func (service *nginx) PostAction(w http.ResponseWriter, r *http.Request) {
	httpresponse.Error(w, http.StatusNotImplemented, "Not implemented")
}
