package ldap

import (
	"context"
	"net/http"
	"regexp"

	// Packages
	server "github.com/mutablelogic/go-server"
	router "github.com/mutablelogic/go-server/pkg/handler/router"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Check interfaces are satisfied
var _ server.ServiceEndpoints = (*ldap)(nil)

// Request to create a user
type reqCreateUser struct {
	Name        string `json:"name"`
	UserId      int    `json:"uid,omitempty"`
	GroupId     int    `json:"gid,omitempty"`
	Description string `json:"description,omitempty"`

	// Other attributes - homeDirectory, loginShell, givenName, surname, mail, etc
	Attrs map[string][]string `json:"attrs,omitempty"`
}

// Request to create a group
type reqCreateGroup struct {
	Name        string `json:"name"`
	GroupId     int    `json:"gid,omitempty"`
	Description string `json:"description,omitempty"`

	// Other attributes
	Attrs map[string][]string `json:"attrs,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	jsonIndent = 2
)

var (
	reUsers  = regexp.MustCompile(`^/u/?$`)
	reUser   = regexp.MustCompile(`^/u/(` + types.ReIdentifier + `)/?$`)
	reGroups = regexp.MustCompile(`^/g/?$`)
	reGroup  = regexp.MustCompile(`^/g/(` + types.ReIdentifier + `)/?$`)
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - ENDPOINTS

// Add endpoints to the router
func (service *ldap) AddEndpoints(ctx context.Context, r server.Router) {
	// Path: /u
	// Methods: GET
	// Scopes: read
	// Description: Return all users
	r.AddHandlerFuncRe(ctx, reUsers, service.reqListUsers, http.MethodGet).(router.Route).
		SetScope(service.ScopeRead()...)

	// Path: /g
	// Methods: GET
	// Scopes: read
	// Description: Return all groups
	r.AddHandlerFuncRe(ctx, reGroups, service.reqListGroups, http.MethodGet).(router.Route).
		SetScope(service.ScopeRead()...)

	/*
		// Path: /u/<name>
		// Methods: GET
		// Scopes: read
		// Description: Return a user
		r.AddHandlerFuncRe(ctx, reUser, service.reqGetUser, http.MethodGet).(router.Route).
			SetScope(service.ScopeRead()...)

		// Path: /g/<name>
		// Methods: GET
		// Scopes: read
		// Description: Return a group
		r.AddHandlerFuncRe(ctx, reGroup, service.reqGetGroup, http.MethodGet).(router.Route).
			SetScope(service.ScopeRead()...)
	*/
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get all users
func (service *ldap) reqListUsers(w http.ResponseWriter, r *http.Request) {
	list, err := service.GetUsers()
	if err != nil {
		httpresponse.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpresponse.JSON(w, list, http.StatusOK, jsonIndent)
}

// Get all groups
func (service *ldap) reqListGroups(w http.ResponseWriter, r *http.Request) {
	list, err := service.GetGroups()
	if err != nil {
		httpresponse.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpresponse.JSON(w, list, http.StatusOK, jsonIndent)
}
