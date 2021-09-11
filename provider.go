package server

import (
	"context"
	"html/template"
	"io"
	"net"
	"net/http"
	"regexp"
	"time"
)

/////////////////////////////////////////////////////////////////////
// PROVIDER INTERFACES

// Provider provides services to a module
type Provider interface {
	Logger
	Router
	EventQueue

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

// Queue allows posting events and subscription to events from other plugins
type EventQueue interface {
	Post(context.Context, Event)
	Subscribe(context.Context, chan<- Event) error
}

/////////////////////////////////////////////////////////////////////
// EVENT INTERFACE

type Event interface {
	Name() string
	Value() interface{}
}

/////////////////////////////////////////////////////////////////////
// SERVICE DISCOVERY INTERFACE

type Service interface {
	Instance() string
	Service() string
	Name() string
	Host() string
	Port() uint16
	Zone() string
	Addrs() []net.IP
	Txt() []string

	// Return TXT keys and value for a key
	Keys() []string
	ValueForKey(string) string
}

/////////////////////////////////////////////////////////////////////
// TEMPLATE & INDEXER INTERFACES

// Renderer translates a data stream into a document
type Renderer interface {
	// Return default mimetypes and file extensions handled by this renderer
	Mimetypes() []string

	// Render a data stream into a document
	Read(context.Context, io.Reader) (Document, error)
}

// Document to be rendered or indexed
type Document interface {
	Title() string
	Description() string
	Shortform() template.HTML
	Tags() []string
	File() DocumentFile
	Meta() map[DocumentKey]interface{}
	HTML() []DocumentSection
}

// DocumentFile represents a document materialized on a file system
type DocumentFile interface {
	// Name of the file, including the file extension
	Name() string

	// Parent path for the file, not including the filename
	Path() string

	// Extension for the file, including the "."
	Ext() string

	// Last modification time for the document
	ModTime() time.Time

	// Size in bytes
	Size() int64
}

// DocumentSection represents document HTML, split into sections
type DocumentSection interface {
	// Title for the section
	Title() string

	// Level of the section, or zero if no level defined
	Level() uint

	// Valid HTML for the section (which allows extraction into text tokens)
	HTML() template.HTML

	// Anchor name for the section, if any
	Anchor() string

	// Class name for the section, if any
	Class() string
}

// DocumentKey provides additional metadata for a document
type DocumentKey string

const (
	DocumentKeyAuthor    DocumentKey = "author"
	DocumentKeyArtwork   DocumentKey = "artwork"
	DocumentKeyThumbnail DocumentKey = "thumbnail"
	DocumentKeyMimetype  DocumentKey = "mimetype"
)
