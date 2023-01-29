package router

import (
	"context"
	"net/http"
	"regexp"

	// Module imports
	ctx "github.com/mutablelogic/go-server/pkg/context"
	util "github.com/mutablelogic/go-server/pkg/httpserver/util"
	plugin "github.com/mutablelogic/go-server/plugin"
)

///////////////////////////////////////////////////////////////////////////////
// SCHEMA

type Gateway struct {
	Prefix      string         `json:"prefix,omitempty"`
	Label       string         `json:"label"`
	Description string         `json:"description,omitempty"`
	Routes      []GatewayRoute `json:"routes,omitempty"`
	Middleware  []string       `json:"middleware,omitempty"`
}

type GatewayRoute struct {
	Path        string   `json:"path"`
	Description string   `json:"description,omitempty"`
	Methods     []string `json:"methods,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
	Middleware  []string `json:"middleware,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	reRoot    = regexp.MustCompile(`^/?$`)
	reGateway = regexp.MustCompile(`^/(.+)/?$`)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Register routes for router
func (gateway *router) RegisterHandlers(parent context.Context, router plugin.Router) {
	// GET /
	//   Return list of prefixes and their handlers
	router.AddHandler(
		ctx.WithDescription(ctx.WithScope(parent, ScopeRead), "Enumerate gateway services"),
		reRoot, gateway.ReqPrefix,
		http.MethodGet,
	)
	// GET /:prefix:
	//   Return list of routes for a prefix
	router.AddHandler(
		ctx.WithDescription(ctx.WithScope(parent, ScopeRead), "Return description of handlers for a gateway service"),
		reGateway, gateway.ReqRoutes,
		http.MethodGet,
	)
}

///////////////////////////////////////////////////////////////////////////////
// HANDLERS

func (router *router) ReqPrefix(w http.ResponseWriter, r *http.Request) {
	util.ServeJSON(w, router.service, http.StatusOK, 2)
}

func (router *router) ReqRoutes(w http.ResponseWriter, r *http.Request) {
	_, _, params := util.ReqPrefixPathParams(r)
	prefix := normalizePath(params[0], false)
	gateway, exists := router.service[prefix]
	if !exists {
		util.ServeError(w, http.StatusNotFound)
	}

	// Build response
	gateway.Prefix = prefix
	for _, route := range router.routes {
		if route.Prefix() != prefix {
			continue
		}
		gateway.Routes = append(gateway.Routes, GatewayRoute{
			Path:        route.Path(),
			Description: route.Description(),
			Methods:     route.Methods(),
			Scopes:      route.Scopes(),
			Middleware:  route.Middleware(),
		})
	}

	// Serve response
	util.ServeJSON(w, gateway, http.StatusOK, 2)
}
