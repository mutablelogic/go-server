package context

import (
	"context"
	"fmt"
	"io"

	// Packages
	hcl "github.com/mutablelogic/go-hcl/pkg/block"
	iface "github.com/mutablelogic/go-server"
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
	contextProvider
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Return a context object with a function to cancel the context
func WithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - SET VALUES IN CONTEXT

// Return a context with the given prefix
func WithProvider(ctx context.Context, provider iface.Provider) context.Context {
	return context.WithValue(ctx, contextProvider, provider)
}

// Return a context with the given prefix
func WithNameLabel(ctx context.Context, name string, label hcl.Label) context.Context {
	return context.WithValue(context.WithValue(ctx, contextName, name), contextLabel, label)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - RETURN VALUES FROM CONTEXT

// Return the name parameter from the context, or zero value if
// not defined
func Name(ctx context.Context) string {
	if value, ok := ctx.Value(contextName).(string); ok {
		return value
	} else {
		return ""
	}
}

// Return the label parameter from the context, or zero value if
// not defined
func Label(ctx context.Context) hcl.Label {
	if value, ok := ctx.Value(contextName).(hcl.Label); ok {
		return value
	} else {
		return nil
	}
}

// Return the name and label combined
func NameLabel(ctx context.Context) hcl.Label {
	if label := Label(ctx); label.IsZero() {
		return hcl.NewLabel(Name(ctx))
	} else {
		return hcl.NewLabel(Name(ctx), label.String())
	}
}

// Return the provider parameter from the context, or zero value if
// not defined
func Provider(ctx context.Context) iface.Provider {
	if value, ok := ctx.Value(contextProvider).(iface.Provider); ok {
		return value
	} else {
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// Stringify the context values to an io.Writer object
func Dump(ctx context.Context, w io.Writer) {
	fmt.Fprintf(w, "<context")
	if value := Name(ctx); value != "" {
		fmt.Fprintf(w, " name=%q", value)
	}
	if value := Label(ctx); value != nil {
		fmt.Fprintf(w, " label=%q", value)
	}
	if value := Provider(ctx); value != nil {
		fmt.Fprintf(w, " provider=%v", value)
	}
	fmt.Fprintf(w, ">")
}
