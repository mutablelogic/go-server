package mdns

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

type RespServices struct {
	Zone     string   `json:"zone,omitempty"`
	Services []string `json:"services,omitempty"`
}

type RespInstances struct {
	Zone      string     `json:"zone,omitempty"`
	Service   string     `json:"service,omitempty"`
	Instances []*Service `json:"instances,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	reRoot     = regexp.MustCompile(`^/?$`)
	reInstance = regexp.MustCompile(`^/(_[A-Za-z0-9\-]+\._(tcp|udp))/?$`)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Register routes for router
func (mdns *mdns) RegisterHandlers(parent context.Context, router plugin.Router) {
	// GET /
	//   Return all services that have been registered
	router.AddHandler(
		ctx.WithDescription(ctx.WithScope(parent, ScopeRead), "Enumerate services"),
		reRoot, mdns.ReqServices,
		http.MethodGet,
	)
	// GET /_service._tcp
	//   Return all instances that are registered for a service
	router.AddHandler(
		ctx.WithDescription(ctx.WithScope(parent, ScopeRead), "Enumerate instances for a service"),
		reInstance, mdns.ReqInstances,
		http.MethodGet,
	)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Label returns the label of the router
func (mdns *mdns) Label() string {
	return mdns.label
}

// Description returns the label of the router
func (mdns *mdns) Description() string {
	return "Registers and resolves mDNS services"
}

///////////////////////////////////////////////////////////////////////////////
// HANDLERS

func (mdns *mdns) ReqServices(w http.ResponseWriter, r *http.Request) {
	resp := RespServices{
		Zone:     mdns.zone,
		Services: mdns.GetNames(),
	}
	util.ServeJSON(w, &resp, http.StatusOK, 2)
}

func (mdns *mdns) ReqInstances(w http.ResponseWriter, r *http.Request) {
	_, _, params := util.ReqPrefixPathParams(r)
	resp := RespInstances{
		Zone:      mdns.zone,
		Service:   params[0],
		Instances: mdns.GetInstances(params[0]),
	}
	util.ServeJSON(w, &resp, http.StatusOK, 2)
}
