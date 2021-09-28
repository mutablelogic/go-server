package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	// Modules
	. "github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/provider"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type logger struct {
	sync.Mutex
}

type handler struct {
	http.HandlerFunc
}

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the module
func New(ctx context.Context, provider Provider) Plugin {
	this := new(logger)
	return this
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *logger) String() string {
	str := "<log"
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PLUGIN

func Name() string {
	return "log"
}

func (this *logger) Run(ctx context.Context, _ Provider) error {
	<-ctx.Done()
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - LOG

func (this *logger) Print(ctx context.Context, v ...interface{}) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	if name := provider.ContextPluginName(ctx); name != "" {
		log.Print("["+name+"] ", fmt.Sprint(v...))
	} else {
		log.Print(v...)
	}
}

func (this *logger) Printf(ctx context.Context, f string, v ...interface{}) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	if name := provider.ContextPluginName(ctx); name != "" {
		log.Print("["+name+"] ", fmt.Sprintf(f, v...))
	} else {
		log.Printf(f, v...)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - MIDDLEWARE

func (this *logger) AddMiddlewareFunc(ctx context.Context, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		this.Print(r.Context(), provider.ContextPluginName(ctx), " ", r.Method, " ", r.URL.Path)
		h(w, r)
	}
}

func (this *logger) AddMiddleware(ctx context.Context, h http.Handler) http.Handler {
	return &handler{
		this.AddMiddlewareFunc(ctx, h.ServeHTTP),
	}
}

func (this *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	this.HandlerFunc(w, r)
}
