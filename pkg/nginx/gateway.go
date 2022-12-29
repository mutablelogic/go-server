package nginx

import (
	"context"
	"net/http"
	"regexp"
	"time"

	// Module imports
	ctx "github.com/mutablelogic/go-server/pkg/context"
	util "github.com/mutablelogic/go-server/pkg/httpserver/util"
	plugin "github.com/mutablelogic/go-server/plugin"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	reRoot   = regexp.MustCompile(`^/?$`)
	reAction = regexp.MustCompile(`^/(test|reload|reopen)/?$`)
)

///////////////////////////////////////////////////////////////////////////////
// SCHEMA

type HealthResponse struct {
	Version string `json:"version"`
	Uptime  uint64 `json:"uptime_secs"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Register routes for router
func (nginx *t) RegisterHandlers(parent context.Context, router plugin.Router) {
	// Path: /
	// Methods: GET
	// Scopes: read
	// Description: Get nginx status (version, uptime, available and enabled configurations)
	router.AddHandler(
		ctx.WithDescription(ctx.WithScope(parent, ScopeRead), "Return nginx version and uptime"),
		reRoot, nginx.ReqHealth,
		http.MethodGet,
	)
	// Path: /(test|reload|reopen)
	// Methods: POST
	// Scopes: write
	// Description: Test, reload and reopen nginx configuration
	router.AddHandler(
		ctx.WithDescription(ctx.WithScope(parent, ScopeWrite), "Test, reload or reopen nginx"),
		reAction, nginx.ReqAction,
		http.MethodPost,
	)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Label returns the label of nginx
func (nginx *t) Label() string {
	return nginx.label
}

// Description returns the label of the router
func (nginx *t) Description() string {
	return "Starts, stops and configures nginx reverse proxy"
}

///////////////////////////////////////////////////////////////////////////////
// HANDLERS

func (nginx *t) ReqHealth(w http.ResponseWriter, r *http.Request) {
	var response HealthResponse

	// Fill response
	response.Version = nginx.Version()
	response.Uptime = uint64(time.Since(nginx.cmd.Start).Seconds())

	// Return response
	util.ServeJSON(w, response, http.StatusOK, 2)
}

func (nginx *t) ReqAction(w http.ResponseWriter, r *http.Request) {
	// Check parameters
	_, _, params := util.ReqPrefixPathParams(r)
	if len(params) != 1 {
		util.ServeError(w, http.StatusInternalServerError)
		return
	}

	switch params[0] {
	case "test":
		if err := nginx.Test(); err != nil {
			util.ServeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	case "reload":
		if err := nginx.Reload(); err != nil {
			util.ServeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	case "reopen":
		if err := nginx.Reopen(); err != nil {
			util.ServeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	default:
		util.ServeError(w, http.StatusNotFound)
		return
	}
}
