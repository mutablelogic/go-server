package provider

import (
	"context"
	"fmt"

	// Modules
	config "github.com/djthorpe/go-server/pkg/config"
)

type contextKey int

const (
	ctxKeyPluginName contextKey = iota
	ctxKeyAddr
	ctxKeyPluginPath
	ctxKeyHandler
	ctxKeyParams
	ctxKeyUser
	ctxKeyAuth
	ctxKeyPath
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
	// Reverse the middleware value
	reverse(handler.Middleware)

	// Return the context
	return context.WithValue(parent, ctxKeyHandler, handler)
}

func ContextWithAuth(parent context.Context, user string, auth map[string]interface{}) context.Context {
	return context.WithValue(context.WithValue(parent, ctxKeyUser, user), ctxKeyAuth, auth)
}

func ContextWithPrefix(ctx context.Context, prefix string) context.Context {
	if handler, ok := ctx.Value(ctxKeyHandler).(config.Handler); !ok {
		return nil
	} else {
		handler.Prefix = prefix
		return ContextWithHandler(ctx, handler)
	}
}

func ContextHandlerPrefix(ctx context.Context) string {
	if handler, ok := ctx.Value(ctxKeyHandler).(config.Handler); !ok {
		return ""
	} else {
		return handler.Prefix
	}
}

func ContextHandlerMiddleware(ctx context.Context) []string {
	if handler, ok := ctx.Value(ctxKeyHandler).(config.Handler); !ok {
		return nil
	} else {
		return handler.Middleware
	}
}

func ContextWithPathParams(parent context.Context, path string, params []string) context.Context {
	return context.WithValue(context.WithValue(parent, ctxKeyParams, params), ctxKeyPath, path)
}

func ContextParams(ctx context.Context) []string {
	if params, ok := ctx.Value(ctxKeyParams).([]string); ok {
		return params
	} else {
		return nil
	}
}

func ContextPath(ctx context.Context) string {
	if path, ok := ctx.Value(ctxKeyPath).(string); ok {
		return path
	} else {
		return ""
	}
}

func ContextUser(ctx context.Context) string {
	if user, ok := ctx.Value(ctxKeyUser).(string); ok {
		return user
	} else {
		return ""
	}
}

func ContextWithAddr(parent context.Context, addr string) context.Context {
	return context.WithValue(parent, ctxKeyAddr, addr)
}

func ContextAddr(ctx context.Context) string {
	if addr, ok := ctx.Value(ctxKeyAddr).(string); ok {
		return addr
	} else {
		return ""
	}
}

func DumpContext(ctx context.Context) string {
	str := "<context"
	if addr, ok := ctx.Value(ctxKeyAddr).(string); ok {
		str += fmt.Sprintf(" addr=%q", addr)
	}
	if name, ok := ctx.Value(ctxKeyPluginName).(string); ok {
		str += fmt.Sprintf(" name=%q", name)
	}
	if path, ok := ctx.Value(ctxKeyPluginPath).([]string); ok {
		str += fmt.Sprintf(" plugin_path=%q", path)
	}
	if handler, ok := ctx.Value(ctxKeyHandler).(config.Handler); ok {
		str += fmt.Sprintf(" handler=%v", handler)
	}
	if path, ok := ctx.Value(ctxKeyPath).(string); ok {
		str += fmt.Sprintf(" req_path=%q", path)
	}
	if params, ok := ctx.Value(ctxKeyParams).([]string); ok {
		str += fmt.Sprintf(" params=%q", params)
	}
	if user, ok := ctx.Value(ctxKeyUser).(string); ok {
		str += fmt.Sprintf(" user=%q", user)
	}
	if auth, ok := ctx.Value(ctxKeyAuth).(map[string]interface{}); ok {
		str += fmt.Sprintf(" auth=%+v", auth)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func reverse(v []string) {
	last := len(v) - 1
	for i := 0; i < len(v)/2; i++ {
		v[i], v[last-i] = v[last-i], v[i]
	}
}
