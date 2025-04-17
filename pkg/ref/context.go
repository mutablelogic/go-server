package ref

import (
	"context"
	"encoding/json"
	"io"

	// Packages
	server "github.com/mutablelogic/go-server"
	authschema "github.com/mutablelogic/go-server/pkg/auth/schema"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ctx int

const (
	ctxProvider ctx = iota
	ctxLogger
	ctxAuth
	ctxUser
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

func Label(ctx context.Context) string {
	if value, ok := ctx.Value(ctxPath).([]string); !ok || len(value) == 0 {
		return ""
	} else {
		return value[len(value)-1]
	}
}

func Log(ctx context.Context) server.Logger {
	if value := ctx.Value(ctxLogger); value == nil {
		return nil
	} else {
		return value.(server.Logger)
	}
}

func Auth(ctx context.Context) server.Auth {
	if value, ok := ctx.Value(ctxAuth).(server.Auth); !ok || value == nil {
		return nil
	} else {
		return value
	}
}

func User(ctx context.Context) *authschema.User {
	if value, ok := ctx.Value(ctxUser).(*authschema.User); !ok || value == nil {
		return nil
	} else {
		return value
	}
}

func WithLog(ctx context.Context, logger server.Logger) context.Context {
	return context.WithValue(ctx, ctxLogger, logger)
}

func WithAuth(ctx context.Context, auth server.Auth) context.Context {
	return context.WithValue(ctx, ctxAuth, auth)
}

func WithUser(ctx context.Context, user *authschema.User) context.Context {
	return context.WithValue(ctx, ctxUser, user)
}

// Set the provider in the context
func WithProvider(parent context.Context, provider server.Provider) context.Context {
	return context.WithValue(context.WithValue(parent, ctxProvider, provider), ctxLogger, provider)
}

// Append a path to the context
func WithPath(parent context.Context, path ...string) context.Context {
	return context.WithValue(parent, ctxPath, append(Path(parent), path...))
}

func DumpContext(w io.Writer, ctx context.Context) error {
	type j struct {
		Path     []string         `json:"path,omitempty"`
		Provider server.Provider  `json:"provider,omitempty"`
		Auth     server.Auth      `json:"auth,omitempty"`
		User     *authschema.User `json:"user,omitempty"`
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(j{
		Path:     Path(ctx),
		Provider: Provider(ctx),
		Auth:     Auth(ctx),
		User:     User(ctx),
	})
}
