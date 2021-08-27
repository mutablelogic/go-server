package main

import (
	"context"
	"net/http"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type handler struct {
	http.HandlerFunc
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - MIDDLEWARE

func (this *plugin) AddMiddlewareFunc(ctx context.Context, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		attributes, expiry := this.validate(w, req)
		if attributes == nil {
			return
		}

		// Set a new token if expired
		if expiry.Before(time.Now()) {
			if err := this.setToken(req.Context(), w, attributes); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Let child do it's thing
		h(w, req)
	}
}

func (this *plugin) AddMiddleware(ctx context.Context, h http.Handler) http.Handler {
	return &handler{
		this.AddMiddlewareFunc(ctx, h.ServeHTTP),
	}
}

func (this *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	this.HandlerFunc(w, r)
}
