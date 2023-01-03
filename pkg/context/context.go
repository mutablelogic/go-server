package context

import (
	"context"
	"fmt"
	"io"
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
	contextAddress
	contextPath
	contextScope
	contextDescription
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Return a context object with a function to cancel the context
func WithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return a context with the given prefix
func WithPrefix(ctx context.Context, prefix string) context.Context {
	return context.WithValue(ctx, contextPrefix, prefix)
}

// Return a context with the given prefix and parameters
func WithPrefixPathParams(ctx context.Context, prefix, path string, params []string) context.Context {
	return context.WithValue(
		context.WithValue(
			context.WithValue(
				ctx, contextParams, params,
			),
			contextPrefix, prefix),
		contextPath, path)
}

// Return a context with the given name
func WithName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, contextName, name)
}

// Return a context with the given name and label
func WithNameLabel(ctx context.Context, name, label string) context.Context {
	return context.WithValue(context.WithValue(ctx, contextName, name), contextLabel, label)
}

// Return a context with the given address string
func WithAddress(ctx context.Context, addr string) context.Context {
	return context.WithValue(ctx, contextAddress, addr)
}

// Return a context with the given path string
func WithPath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, contextPath, path)
}

// Return a context with the given set of scopes
func WithScope(ctx context.Context, scope ...string) context.Context {
	if len(scope) > 0 {
		return context.WithValue(ctx, contextScope, scope)
	} else {
		return ctx
	}
}

// Return a context with a description
func WithDescription(ctx context.Context, description string) context.Context {
	return context.WithValue(ctx, contextDescription, description)
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

// Return the name and label parameter from the context, or zero value if
// not defined
func NameLabel(ctx context.Context) string {
	return Name(ctx) + "." + Label(ctx)
}

// Return the address parameter from the context, or zero value if
// not defined
func Address(ctx context.Context) string {
	return contextString(ctx, contextAddress)
}

// Return the path parameter from the context, or zero value if
// not defined
func Path(ctx context.Context) string {
	return contextString(ctx, contextPath)
}

// Return the prefix parameter from the context, or zero value if
// not defined
func Prefix(ctx context.Context) string {
	return contextString(ctx, contextPrefix)
}

// Return prefix and parameters from the context
func PrefixPathParams(ctx context.Context) (string, string, []string) {
	return contextString(ctx, contextPrefix), contextString(ctx, contextPath), contextStringSlice(ctx, contextParams)
}

// Return array of scopes from the context, or nil
func Scope(ctx context.Context) []string {
	return contextStringSlice(ctx, contextScope)
}

// Return description from the context, zero value if not defined
func Description(ctx context.Context) string {
	return contextString(ctx, contextDescription)
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
	if value, ok := ctx.Value(contextPath).(string); ok {
		fmt.Fprintf(w, " path=%q", value)
	}
	if value, ok := ctx.Value(contextParams).([]string); ok {
		fmt.Fprintf(w, " params=%q", value)
	}
	if value, ok := ctx.Value(contextScope).([]string); ok {
		fmt.Fprintf(w, " scope=%q", value)
	}
	if value, ok := ctx.Value(contextDescription).(string); ok {
		fmt.Fprintf(w, " description=%q", value)
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

/*
func contextBool(ctx context.Context, key contextType) bool {
	if value, ok := ctx.Value(key).(bool); ok {
		return value
	} else {
		return false
	}
}
*/

func contextStringSlice(ctx context.Context, key contextType) []string {
	if value, ok := ctx.Value(key).([]string); ok {
		return value
	} else {
		return nil
	}
}
