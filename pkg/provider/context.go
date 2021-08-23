package provider

import (
	"context"
	"fmt"

	config "github.com/djthorpe/go-server/pkg/config"
)

type contextKey int

const (
	ctxKeyPluginName contextKey = iota
	ctxKeyPluginPath
	ctxKeyHandler
	ctxKeyParams
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func ContextWithPluginName(ctx context.Context, name string) context.Context {
	parent := ContextPluginName(ctx)
	ctx = context.WithValue(ctx, ctxKeyPluginName, name)
	if path, ok := ctx.Value(ctxKeyPluginPath).([]string); !ok {
		if parent == "" {
			return context.WithValue(ctx, ctxKeyPluginPath, []string{})
		} else {
			return context.WithValue(ctx, ctxKeyPluginPath, []string{parent})
		}
	} else {
		return context.WithValue(ctx, ctxKeyPluginPath, append(path, parent))
	}
}

func ContextHasPluginParent(ctx context.Context, v string) bool {
	path, _ := ctx.Value(ctxKeyPluginPath).([]string)
	for _, name := range path {
		if name == v {
			return true
		}
	}
	return false
}

func ContextPluginName(ctx context.Context) string {
	if name, ok := ctx.Value(ctxKeyPluginName).(string); ok {
		return name
	} else {
		return ""
	}
}

func ContextWithHandler(parent context.Context, handler config.Handler) context.Context {
	return context.WithValue(parent, ctxKeyHandler, handler)
}

func ContextHandlerPrefix(ctx context.Context) string {
	if handler, ok := ctx.Value(ctxKeyHandler).(config.Handler); !ok {
		return ""
	} else {
		return handler.Prefix
	}
}

func ContextWithParams(parent context.Context, params []string) context.Context {
	return context.WithValue(parent, ctxKeyParams, params)
}

func ContextParams(ctx context.Context) []string {
	if params, ok := ctx.Value(ctxKeyParams).([]string); ok {
		return params
	} else {
		return nil
	}
}

func DumpContext(ctx context.Context) string {
	str := "<context"
	if name, ok := ctx.Value(ctxKeyPluginName).(string); ok {
		str += fmt.Sprintf(" name=%q", name)
	}
	if path, ok := ctx.Value(ctxKeyPluginPath).([]string); ok {
		str += fmt.Sprintf(" path=%q", path)
	}
	if handler, ok := ctx.Value(ctxKeyHandler).(config.Handler); ok {
		str += fmt.Sprintf(" handler=%v", handler)
	}
	if params, ok := ctx.Value(ctxKeyParams).([]string); ok {
		str += fmt.Sprintf(" params=%q", params)
	}
	return str + ">"
}
