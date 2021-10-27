package main

import (
	"context"
	"net/http"
	"net/url"
	"regexp"

	// Packages
	router "github.com/mutablelogic/go-server/pkg/httprouter"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type PingResponse struct {
	Users  []url.Values `json:"users"`
	Groups []url.Values `json:"groups"`
}

///////////////////////////////////////////////////////////////////////////////
// ROUTES

var (
	reRoutePing = regexp.MustCompile(`^/?$`)
)

const (
	defaultLimit = 100
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (p *plugin) addHandlers(ctx context.Context, provider Provider) error {
	// Add handler for returning OK when ldap server can be reached
	if err := provider.AddHandlerFuncEx(ctx, reRoutePing, p.ServePing); err != nil {
		provider.Print(ctx, "Failed to add handler: ", err)
		return nil
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// HANDLERS

func (p *plugin) ServePing(w http.ResponseWriter, req *http.Request) {
	if err := p.Ping(); err != nil {
		router.ServeError(w, http.StatusBadGateway, err.Error())
		return
	}

	var response PingResponse

	// Get all users
	if results, err := p.ListUsers(defaultLimit); err != nil {
		router.ServeError(w, http.StatusBadRequest, err.Error())
		return
	} else {
		response.Users = results
	}

	// Get all groups
	if results, err := p.ListGroups(defaultLimit); err != nil {
		router.ServeError(w, http.StatusBadRequest, err.Error())
		return
	} else {
		response.Groups = results
	}

	// Return response
	router.ServeJSON(w, response, http.StatusOK, 2)
}
