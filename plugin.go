package server

import (
	"context"
	"io"
	"io/fs"
	"net/http"
	"net/url"

	// Packages
	pg "github.com/djthorpe/go-pg"
	authschema "github.com/mutablelogic/go-server/pkg/auth/schema"
	ldapschema "github.com/mutablelogic/go-server/pkg/ldap/schema"
)

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

// Plugin represents a service
type Plugin interface {
	// Return the unique name for the plugin
	Name() string

	// Return a description of the plugin
	Description() string

	// Create a task from a plugin
	New(context.Context) (Task, error)
}

// Task represents a service task
type Task interface {
	// Run a task
	Run(context.Context) error
}

// Provider represents a service provider
type Provider interface {
	Task

	// Write out the plugin configuration to a writer
	WriteConfig(io.Writer) error

	// Return the plugins registered with the provider
	Plugins() []Plugin

	// Load a plugin by name and label, and provide a resolver function
	Load(string, string, PluginResolverFunc) error

	// Return a task given a plugin label
	Task(context.Context, string) Task
}

// Define a function for resolving plugin configuration
type PluginResolverFunc func(context.Context, string, Plugin) error

///////////////////////////////////////////////////////////////////////////////
// HTTP ROUTER

type HTTPRouter interface {
	// Return the CORS origin
	Origin() string

	// Register a function to handle a URL path
	HandleFunc(context.Context, string, http.HandlerFunc)

	// Register serving of static files from a filesystem
	HandleFS(context.Context, string, fs.FS)
}

///////////////////////////////////////////////////////////////////////////////
// HTTP MIDDLEWARE

type HTTPMiddleware interface {
	// Return a http handler with middleware as the parent handler
	HandleFunc(http.HandlerFunc) http.HandlerFunc
}

///////////////////////////////////////////////////////////////////////////////
// LOGGING AND METRICS

// Logger defines methods for logging messages and structured data.
// It can also act as HTTP middleware for request logging.
type Logger interface {
	HTTPMiddleware

	// Debug logs a debugging message.
	Debug(context.Context, ...any)

	// Debugf logs a formatted debugging message.
	Debugf(context.Context, string, ...any)

	// Print logs an informational message.
	Print(context.Context, ...any)

	// Printf logs a formatted informational message.
	Printf(context.Context, string, ...any)

	// With returns a new Logger that includes the given key-value pairs
	// in its structured log output.
	With(...any) Logger
}

///////////////////////////////////////////////////////////////////////////////
// PGPOOL

// PG provides access to a PostgreSQL connection pool.
type PG interface {
	// Conn returns the underlying connection pool object.
	Conn() pg.PoolConn
}

///////////////////////////////////////////////////////////////////////////////
// AUTHENTICATION AND AUTHORIZATION

// Auth defines methods for authenticating users based on HTTP requests
// and authorizing them based on required scopes.
type Auth interface {
	// Authenticate attempts to identify and return the user associated with
	// an incoming HTTP request, typically by inspecting headers or tokens.
	Authenticate(*http.Request) (*authschema.User, error)

	// Authorize checks if a given user has the necessary permissions (scopes)
	// to perform an action.
	Authorize(context.Context, *authschema.User, ...string) error
}

///////////////////////////////////////////////////////////////////////////////
// LDAP

type LDAP interface {
	// Introspection
	ListObjectClasses(context.Context) ([]*ldapschema.ObjectClass, error)    // Return all classes
	ListAttributeTypes(context.Context) ([]*ldapschema.AttributeType, error) // Return all attributes

	// Objects
	List(context.Context, ldapschema.ObjectListRequest) (*ldapschema.ObjectList, error) // List all objects in the directory
	Get(context.Context, string) (*ldapschema.Object, error)                            // Get an object
	Create(context.Context, string, url.Values) (*ldapschema.Object, error)             // Create a new object with attributes
	Update(context.Context, string, url.Values) (*ldapschema.Object, error)             // Update an object with attributes
	Delete(context.Context, string) (*ldapschema.Object, error)                         // Delete an object

	// Users
	ListUsers(context.Context, ldapschema.ObjectListRequest) (*ldapschema.ObjectList, error) // List users
	GetUser(context.Context, string) (*ldapschema.Object, error)                             // Get a user
	CreateUser(context.Context, string, url.Values) (*ldapschema.Object, error)              // Create a user with attributes
	DeleteUser(context.Context, string) (*ldapschema.Object, error)                          // Delete a user
	// UpdateUser(context.Context, string, url.Values) (*ldapschema.Object, error) // Update a user with attributes

	// Groups
	ListGroups(context.Context, ldapschema.ObjectListRequest) (*ldapschema.ObjectList, error) // List groups
	GetGroup(context.Context, string) (*ldapschema.Object, error)                             // Get a group
	CreateGroup(context.Context, string, url.Values) (*ldapschema.Object, error)              // Create a group with attributes
	DeleteGroup(context.Context, string) (*ldapschema.Object, error)                          // Delete a group
	// UpdateGroup(context.Context, string, url.Values) (*ldapschema.Object, error) // Update a group with attributes
	// AddGroupUser(context.Context, string, string) (*ldapschema.Object, error)                   // Add a user to a group
	// RemoveGroupUser(context.Context, string, string) (*ldapschema.Object, error)                // Remove a user from a group

	// Auth
	Bind(context.Context, string, string) (*ldapschema.Object, error)                    // Check credentials
	ChangePassword(context.Context, string, string, *string) (*ldapschema.Object, error) // Change password for a user, and return the user object
}
