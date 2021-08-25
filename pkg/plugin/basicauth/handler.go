package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	// Modules
	. "github.com/djthorpe/go-server"
	"github.com/djthorpe/go-server/pkg/htpasswd"
	router "github.com/djthorpe/go-server/pkg/httprouter"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type PasswordRequest struct {
	Password string `json:"password"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	reRouteUsers = regexp.MustCompile(`^/?$`)
	reRouteUser  = regexp.MustCompile(`^/(\w+)$`)
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

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// HANDLERS

func (this *basicauth) ServeUsers(w http.ResponseWriter, req *http.Request) {
	response := []string{}

	// Get all users
	if this.Htpasswd != nil {
		this.Mutex.Lock()
		defer this.Mutex.Unlock()
		response = append(response, this.Htpasswd.Users()...)
	}

	// Response
	router.ServeJSON(w, response, http.StatusOK, 2)
}

func (this *basicauth) AddUser(w http.ResponseWriter, req *http.Request) {
	var request PasswordRequest

	params := router.RequestParams(req)
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
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
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
	params := router.RequestParams(req)

	// Check
	if this.Htpasswd == nil {
		router.ServeError(w, http.StatusInternalServerError)
		return
	}

	// Delete password
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	this.Htpasswd.Delete(params[0])
	this.dirty = true

	// Response
	router.ServeError(w, http.StatusOK, fmt.Sprintf("Deleted %q", params[0]))
}
