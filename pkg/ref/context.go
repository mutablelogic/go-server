package ref

import (
	"context"
	"encoding/json"
	"io"

	// Packages
	server "github.com/mutablelogic/go-server"
	auth "github.com/mutablelogic/go-server/pkg/auth/schema"
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
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
	ctxTicker
	ctxTask
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
	if provider := ctx.Value(ctxProvider); provider == nil {
		return nil
	} else {
		return provider.(server.Logger)
	}
}

func Auth(ctx context.Context) server.Auth {
	if value, ok := ctx.Value(ctxAuth).(server.Auth); !ok || value == nil {
		return nil
	} else {
		return value
	}
}

func User(ctx context.Context) *auth.User {
	if value, ok := ctx.Value(ctxUser).(*auth.User); !ok || value == nil {
		return nil
	} else {
		return value
	}
}

func Ticker(ctx context.Context) *pgqueue.Ticker {
	if value, ok := ctx.Value(ctxTicker).(*pgqueue.Ticker); !ok || value == nil {
		return nil
	} else {
		return value
	}
}

func Task(ctx context.Context) *pgqueue.Task {
	if value, ok := ctx.Value(ctxTask).(*pgqueue.Task); !ok || value == nil {
		return nil
	} else {
		return value
	}
}

func WithAuth(ctx context.Context, auth server.Auth) context.Context {
	return context.WithValue(ctx, ctxAuth, auth)
}

func WithUser(ctx context.Context, user *auth.User) context.Context {
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

// Set the ticker in the context
func WithTicker(parent context.Context, ticker *pgqueue.Ticker) context.Context {
	return context.WithValue(parent, ctxTicker, ticker)
}

// Set the task in the context
func WithTask(parent context.Context, task *pgqueue.Task) context.Context {
	return context.WithValue(parent, ctxTask, task)
}

func DumpContext(w io.Writer, ctx context.Context) error {
	type j struct {
		Path     []string        `json:"path,omitempty"`
		Provider server.Provider `json:"provider,omitempty"`
		Auth     server.Auth     `json:"auth,omitempty"`
		User     *auth.User      `json:"user,omitempty"`
		Ticker   *pgqueue.Ticker `json:"ticker,omitempty"`
		Task     *pgqueue.Task   `json:"task,omitempty"`
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(j{
		Path:     Path(ctx),
		Provider: Provider(ctx),
		Auth:     Auth(ctx),
		User:     User(ctx),
		Ticker:   Ticker(ctx),
		Task:     Task(ctx),
	})
}
