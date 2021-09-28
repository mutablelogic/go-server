package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	// Modules
	htpasswd "github.com/mutablelogic/go-server/pkg/htpasswd"
	router "github.com/mutablelogic/go-server/pkg/httprouter"
	provider "github.com/mutablelogic/go-server/pkg/provider"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type UsersResponse struct {
	Users  []string        `json:"users"`
	Groups []GroupResponse `json:"groups,omitempty"`
}

type GroupResponse struct {
	Name    string   `json:"name"`
	Members []string `json:"members,omitempty"`
}

type PasswordRequest struct {
	Password string `json:"password"`
}

type GroupRequest struct {
	User string `json:"user"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	reRouteUsers = regexp.MustCompile(`^/?$`)
	reRouteUser  = regexp.MustCompile(`^/(\w+)$`)
	reRouteGroup = regexp.MustCompile(`^/` + htpasswd.GroupPrefix + `(\w+)$`)
)

///////////////////////////////////////////////////////////////////////////////
// ADD HANDLERS

func (this *basicauth) addHandlers(ctx context.Context, provider Provider) error {
	// Add handler for users
	if err := provider.AddHandlerFuncEx(ctx, reRouteUsers, this.ServeUsers); err != nil {
		return err
	}

	// Add handler to create user
	if err := provider.AddHandlerFuncEx(ctx, reRouteUser, this.AddUser, http.MethodPut, http.MethodPost); err != nil {
		return err
	}

	// Add handler to delete user
	if err := provider.AddHandlerFuncEx(ctx, reRouteUser, this.DeleteUser, http.MethodDelete); err != nil {
		return err
	}

	// Add handler to add user to group
	if err := provider.AddHandlerFuncEx(ctx, reRouteGroup, this.AddGroup, http.MethodPut, http.MethodPost); err != nil {
		return err
	}

	// Add handler to remove user from group
	if err := provider.AddHandlerFuncEx(ctx, reRouteGroup, this.DeleteGroup, http.MethodDelete); err != nil {
		return err
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// HANDLERS

func (this *basicauth) AuthorizeRequest(req *http.Request) bool {
	// When no admin user is set, allow all requests
	if this.Config.Admin == "" {
		return true
	}

	// Check for group name
	user := provider.ContextUser(req.Context())
	if strings.HasPrefix(this.Config.Admin, htpasswd.GroupPrefix) {
		group := strings.TrimPrefix(this.Config.Admin, htpasswd.GroupPrefix)
		if this.Htgroups != nil {
			return this.Htgroups.UserInGroup(user, group)
		} else {
			return false
		}
	}

	// Check for user name
	return user == this.Config.Admin
}

func (this *basicauth) ServeUsers(w http.ResponseWriter, req *http.Request) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Authorize request
	if !this.AuthorizeRequest(req) {
		router.ServeError(w, http.StatusUnauthorized)
		return
	}

	response := UsersResponse{
		Users:  []string{},
		Groups: []GroupResponse{},
	}
	if this.Htpasswd != nil {
		response.Users = this.Htpasswd.Users()
	}
	if this.Htgroups != nil {
		for _, group := range this.Htgroups.Groups() {
			response.Groups = append(response.Groups, GroupResponse{
				Name:    htpasswd.GroupPrefix + group,
				Members: this.Htgroups.UsersForGroup(group),
			})
		}
	}

	// Response
	router.ServeJSON(w, response, http.StatusOK, 2)
}

func (this *basicauth) AddUser(w http.ResponseWriter, req *http.Request) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Authorize request
	if !this.AuthorizeRequest(req) {
		router.ServeError(w, http.StatusUnauthorized)
		return
	}

	// Decode request body
	var request PasswordRequest
	if err := router.RequestBody(req, &request); err != nil {
		router.ServeError(w, http.StatusBadRequest)
		return
	} else if request.Password == "" {
		router.ServeError(w, http.StatusBadRequest, "Missing password")
		return
	}

	// Check
	if this.Htpasswd == nil {
		router.ServeError(w, http.StatusInternalServerError)
		return
	}

	// Set password
	params := router.RequestParams(req)
	if err := this.Htpasswd.Set(params[0], request.Password, htpasswd.BCrypt); err != nil {
		router.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	} else {
		this.dirty = true
	}

	// Response
	router.ServeError(w, http.StatusCreated, fmt.Sprintf("Added or modified %q", params[0]))
}

func (this *basicauth) DeleteUser(w http.ResponseWriter, req *http.Request) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Authorize request
	if !this.AuthorizeRequest(req) {
		router.ServeError(w, http.StatusUnauthorized)
		return
	}

	// Check
	if this.Htpasswd == nil {
		router.ServeError(w, http.StatusInternalServerError)
		return
	}

	// Remove user from any groups
	params := router.RequestParams(req)
	if this.Htgroups != nil {
		for _, group := range this.Htgroups.GroupsForUser(params[0]) {
			this.Htgroups.RemoveUserFromGroup(params[0], group)
		}
	}

	// Delete password
	this.Htpasswd.Delete(params[0])
	this.dirty = true

	// Response
	router.ServeError(w, http.StatusOK, fmt.Sprintf("Deleted %q", params[0]))
}

func (this *basicauth) AddGroup(w http.ResponseWriter, req *http.Request) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Authorize request
	if !this.AuthorizeRequest(req) {
		router.ServeError(w, http.StatusUnauthorized)
		return
	}

	// Decode request body
	var request GroupRequest
	if err := router.RequestBody(req, &request); err != nil {
		router.ServeError(w, http.StatusBadRequest)
		return
	} else if request.User == "" {
		router.ServeError(w, http.StatusBadRequest, "Missing user")
		return
	}

	// Check groups and password objects
	if this.Htgroups == nil || this.Htpasswd == nil {
		router.ServeError(w, http.StatusInternalServerError)
		return
	}

	// Check to make sure user exists
	if !this.Htpasswd.Exists(request.User) {
		router.ServeError(w, http.StatusNotFound, "User not found")
		return
	}

	// Add user to groups
	params := router.RequestParams(req)
	if err := this.Htgroups.AddUserToGroup(request.User, params[0]); err != nil {
		router.ServeError(w, http.StatusBadRequest, err.Error())
		return
	} else {
		this.dirty = true
	}

	// Response
	router.ServeError(w, http.StatusCreated, fmt.Sprintf("Added or modified %q", params[0]))
}

func (this *basicauth) DeleteGroup(w http.ResponseWriter, req *http.Request) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Authorize request
	if !this.AuthorizeRequest(req) {
		router.ServeError(w, http.StatusUnauthorized)
		return
	}

	// Decode request query
	var request GroupRequest
	if err := router.RequestQuery(req, &request); err != nil {
		router.ServeError(w, http.StatusBadRequest)
		return
	} else if request.User == "" {
		router.ServeError(w, http.StatusBadRequest, "Missing user")
		return
	}

	// Check groups and password objects
	if this.Htgroups == nil || this.Htpasswd == nil {
		router.ServeError(w, http.StatusInternalServerError)
		return
	}

	// Check user is in the group
	params := router.RequestParams(req)
	if !this.Htgroups.UserInGroup(request.User, params[0]) {
		router.ServeError(w, http.StatusNotFound, "User not found")
		return
	}

	// Remove user from groups
	if err := this.Htgroups.RemoveUserFromGroup(request.User, params[0]); err != nil {
		router.ServeError(w, http.StatusBadRequest, err.Error())
		return
	} else {
		this.dirty = true
	}

	// Response
	router.ServeError(w, http.StatusOK, fmt.Sprintf("Deleted %q", params[0]))
}
