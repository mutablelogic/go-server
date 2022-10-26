package context

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type contextType uint

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	contextNone contextType = iota
	contextName
	contextLabel
	contextPrefix
	contextParams
	contextAdmin
	contextAddress
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Return a context object with a function to cancel the context
func WithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return a context with the given prefix and parameters
func WithPrefixParams(ctx context.Context, prefix string, params []string) context.Context {
	return context.WithValue(context.WithValue(ctx, contextParams, params), contextPrefix, prefix)
}

// Return a context with the given name
func WithName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, contextName, name)
}

// Return a context with the given name and label
func WithNameLabel(ctx context.Context, name, label string) context.Context {
	return context.WithValue(context.WithValue(ctx, contextName, name), contextLabel, label)
}

// Return a context with the given admin flag
func WithAdmin(ctx context.Context, admin bool) context.Context {
	return context.WithValue(ctx, contextAdmin, admin)
}

// Return a context with the given address string
func WithAddress(ctx context.Context, addr string) context.Context {
	return context.WithValue(ctx, contextAddress, addr)
}

///////////////////////////////////////////////////////////////////////////////
// RETURN VALUES FROM CONTEXT

// Return the name parameter from the context, or zero value if
// not defined
func Name(ctx context.Context) string {
	return contextString(ctx, contextName)
}

// Return the label parameter from the context, or zero value if
// not defined
func Label(ctx context.Context) string {
	return contextString(ctx, contextLabel)
}

// Return the admin parameter from the context, or zero value if
// not defined
func Admin(ctx context.Context) bool {
	return contextBool(ctx, contextAdmin)
}

// Return the address parameter from the context, or zero value if
// not defined
func Address(ctx context.Context) string {
	return contextString(ctx, contextAddress)
}

// Return the parameters from a HTTP request, or nil if
// not defined
func ReqParams(req *http.Request) []string {
	if value, ok := req.Context().Value(contextParams).([]string); ok {
		return value
	} else {
		return nil
	}
}

// Return the prefix parameter from a HTTP request, or zero value if
// not defined
func ReqPrefix(req *http.Request) string {
	return contextString(req.Context(), contextPrefix)
}

// Return the name parameter from a HTTP request, or zero value if
// not defined
func ReqName(req *http.Request) string {
	return contextString(req.Context(), contextName)
}

// Return the admin parameter from a HTTP request, or zero value if
// not defined
func ReqAdmin(req *http.Request) bool {
	return contextBool(req.Context(), contextAdmin)
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// Stringify the context values to an io.Writer object
func DumpContext(ctx context.Context, w io.Writer) {
	fmt.Fprintf(w, "<context")
	if value, ok := ctx.Value(contextName).(string); ok {
		fmt.Fprintf(w, " name=%q", value)
	}
	if value, ok := ctx.Value(contextLabel).(string); ok {
		fmt.Fprintf(w, " label=%q", value)
	}
	if value, ok := ctx.Value(contextPrefix).(string); ok {
		fmt.Fprintf(w, " prefix=%q", value)
	}
	if value, ok := ctx.Value(contextParams).([]string); ok {
		fmt.Fprintf(w, " params=%q", value)
	}
	if value, ok := ctx.Value(contextBool).(bool); ok {
		fmt.Fprintf(w, " admin=%v", value)
	}
	if value, ok := ctx.Value(contextAddress).(string); ok {
		fmt.Fprintf(w, " address=%q", value)
	}
	fmt.Fprintf(w, ">")
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func contextString(ctx context.Context, key contextType) string {
	if value, ok := ctx.Value(key).(string); ok {
		return value
	} else {
		return ""
	}
}

func contextBool(ctx context.Context, key contextType) bool {
	if value, ok := ctx.Value(key).(bool); ok {
		return value
	} else {
		return false
	}
}
