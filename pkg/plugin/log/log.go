package main

import (
	"context"
	"log"
	"net/http"
	"sync"

	// Modules
	. "github.com/djthorpe/go-server"
	"github.com/djthorpe/go-server/pkg/provider"
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

func (this *logger) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - LOG

func (this *logger) Print(_ context.Context, v ...interface{}) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	log.Print(v...)
}

func (this *logger) Printf(_ context.Context, fmt string, v ...interface{}) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	log.Printf(fmt, v...)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - MIDDLEWARE

func (this *logger) AddHandlerFunc(ctx context.Context, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		this.Print(r.Context(), provider.ContextPluginName(ctx), " ", r.Method, " ", r.URL.Path)
		h(w, r)
	}
}

func (this *logger) AddHandler(ctx context.Context, h http.Handler) http.Handler {
	return &handler{
		this.AddHandlerFunc(ctx, h.ServeHTTP),
	}
}

func (this *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	this.HandlerFunc(w, r)
}
