package provider

import (
	"context"

	// Packages
	server "github.com/mutablelogic/go-server"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ctx int

const (
	ctxProvider ctx = iota
	ctxPath
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Provider(ctx context.Context) server.Provider {
	if value := ctx.Value(ctxProvider); value == nil {
		return nil
	} else {
		return value.(server.Provider)
	}
}

func Path(ctx context.Context) []string {
	if value, ok := ctx.Value(ctxPath).([]string); !ok {
		return nil
	} else {
		return value
	}
}

func Log(ctx context.Context) server.Logger {
	if value := ctx.Value(ctxProvider); value == nil {
		return nil
	} else {
		return value.(server.Logger)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Set the provider in the context
func withProvider(parent context.Context, provider server.Provider) context.Context {
	return context.WithValue(parent, ctxProvider, provider)
}

// Append a path to the context
func withPath(parent context.Context, path ...string) context.Context {
	return context.WithValue(parent, ctxPath, append(Path(parent), path...))
}
