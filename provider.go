package server

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"time"
)

// Provider provides services to a module
type Provider interface {
	Logger
	Router

	// Plugins returns a list of registered plugin names
	Plugins() []string

	// GetPlugin returns a named plugin or nil if not available
	GetPlugin(context.Context, string) Plugin

	// GetConfig populates yaml config
	GetConfig(context.Context, interface{}) error
}

// Plugin provides handlers to server
type Plugin interface {
	// Run plugin background tasks until cancelled
	Run(context.Context, Provider) error
}

// Logger providers a logging interface
type Logger interface {
	Print(context.Context, ...interface{})
	Printf(context.Context, string, ...interface{})
}

// Router allows handlers to be added for serving URL paths
type Router interface {
	AddHandler(context.Context, http.Handler, ...string) error
	AddHandlerFunc(context.Context, http.HandlerFunc, ...string) error
	AddHandlerFuncEx(context.Context, *regexp.Regexp, http.HandlerFunc, ...string) error
}

// Middleware intercepts HTTP requests
type Middleware interface {
	// Add a child handler object which intercepts a handler
	AddMiddleware(context.Context, http.Handler) http.Handler

	// Add a child handler function which intercepts a handler function
	AddMiddlewareFunc(context.Context, http.HandlerFunc) http.HandlerFunc
}

// Renderer translates a data stream into a document
type Renderer interface {
	Read(context.Context, io.Reader) (Document, error)
}

// Document represents a document to be presented or indexed
type Document struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Author      string   `json:"author,omitempty"`
	Shortform   string   `json:"shortform,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	File        struct {
		Name    string    `json:"name"`
		Path    string    `json:"path"`
		Ext     string    `json:"ext"`
		ModTime time.Time `json:"modtime,omitempty"`
		Size    int64     `json:"size,omitempty"`
	} `json:"file,omitempty"`
	Meta map[DocumentKey]interface{} `json:"meta,omitempty"`
}

// DocumentKey provides additional metadata for a document
type DocumentKey string

const (
	DocumentKeyArtwork   DocumentKey = "artwork"
	DocumentKeyThumbnail DocumentKey = "thumbnail"
)
