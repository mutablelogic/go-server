package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	// Packages
	router "github.com/mutablelogic/go-server/pkg/httprouter"
	provider "github.com/mutablelogic/go-server/pkg/provider"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - MIDDLEWARE

func (this *basicauth) AddMiddlewareFunc(ctx context.Context, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var auth bool
		var user string
		if token := r.Header.Get("Authorization"); token != "" {
			if basic := reBasicAuth.FindStringSubmatch(token); basic != nil {
				basic[1] = strings.TrimRight(basic[1], "=")
				if credentials, err := base64.RawStdEncoding.DecodeString(basic[1]); err == nil {
					user, auth = this.authenticated(credentials)
				}
			}
		}
		// Return authentication error
		if !auth {
			w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=%q, charset=%q", this.Realm, defaultCharset))
			router.ServeError(w, http.StatusUnauthorized)
			return
		}

		// Add user, groups and realm to context
		ctx := r.Context()
		if user != "" {
			ctx = provider.ContextWithAuth(ctx, user, map[string]interface{}{
				"realm":  this.Realm,
				"groups": this.groupsForUser(user),
			})
		}

		// Handle
		h(w, r.Clone(ctx))
	}
}

func (this *basicauth) AddMiddleware(ctx context.Context, h http.Handler) http.Handler {
	return &handler{
		this.AddMiddlewareFunc(ctx, h.ServeHTTP),
	}
}

func (this *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	this.HandlerFunc(w, r)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// authenticated the user and returns true if the given credentials are authenticated
func (this *basicauth) authenticated(credentials []byte) (string, bool) {
	// Lock for reading
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	if creds := strings.SplitN(string(credentials), ":", 2); len(creds) != 2 {
		return "", false
	} else if this.Htpasswd == nil {
		return creds[0], false
	} else {
		return creds[0], this.Htpasswd.Verify(creds[0], creds[1])
	}
}

// groupsForUser returns the groups for the given user
func (this *basicauth) groupsForUser(user string) []string {
	// Lock for reading
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	if this.Htgroups == nil {
		return nil
	} else {
		return this.Htgroups.GroupsForUser(user)
	}
}
