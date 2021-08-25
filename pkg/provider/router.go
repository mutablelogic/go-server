package provider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	// Modules
	multierror "github.com/hashicorp/go-multierror"
)

///////////////////////////////////////////////////////////////////////////////
// ROUTER IMPLEMENTATION

func (this *provider) AddHandler(ctx context.Context, handler http.Handler, methods ...string) error {
	if prefix := ContextHandlerPrefix(ctx); prefix == "" {
		this.Printf(ctx, "AddHandler: Not adding handler for %q, no handler prefix", ContextPluginName(ctx))
		return nil
	} else if len(this.routers) == 0 {
		this.Printf(ctx, "AddHandler: Not adding handler for %q, no router", ContextPluginName(ctx))
		return nil
	}

	// Set middleware
	for _, name := range ContextHandlerMiddleware(ctx) {
		if middleware, exists := this.middleware[name]; !exists {
			return fmt.Errorf("AddMiddleware %q not found for plugin %q", name, ContextPluginName(ctx))
		} else {
			handler = middleware.AddMiddleware(ctx, handler)
		}
	}

	// Log adding handler
	this.Printf(ctx, "Adding Handler %v => prefix=%q methods=%q middleware=%q", ContextPluginName(ctx), ContextHandlerPrefix(ctx), methods, ContextHandlerMiddleware(ctx))

	// Set handler in all routers
	var result error
	for _, router := range this.routers {
		if err := router.AddHandler(ctx, handler, methods...); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Return any errors
	return result
}

func (this *provider) AddHandlerFunc(ctx context.Context, handler http.HandlerFunc, methods ...string) error {
	if prefix := ContextHandlerPrefix(ctx); prefix == "" {
		this.Printf(ctx, "AddHandlerFunc: Not adding handler for %q, no handler prefix", ContextPluginName(ctx))
		return nil
	} else if len(this.routers) == 0 {
		this.Printf(ctx, "AddHandlerFunc: Not adding handler for %q, no router", ContextPluginName(ctx))
		return nil
	}

	// Set middleware, in reverse order
	for _, name := range ContextHandlerMiddleware(ctx) {
		if middleware, exists := this.middleware[name]; !exists {
			return fmt.Errorf("AddMiddleware %q not found for plugin %q", name, ContextPluginName(ctx))
		} else {
			handler = middleware.AddMiddlewareFunc(ctx, handler)
		}
	}

	// Log adding handler
	this.Printf(ctx, "Adding Handler %v => prefix=%q methods=%q middleware=%q", ContextPluginName(ctx), ContextHandlerPrefix(ctx), methods, ContextHandlerMiddleware(ctx))

	// Set handler in all routers
	var result error
	for _, router := range this.routers {
		if err := router.AddHandlerFunc(ctx, handler, methods...); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Return any errors
	return result
}

func (this *provider) AddHandlerFuncEx(ctx context.Context, re *regexp.Regexp, handler http.HandlerFunc, methods ...string) error {
	if prefix := ContextHandlerPrefix(ctx); prefix == "" {
		this.Printf(ctx, "AddHandlerFuncEx: Not adding handler for %q, no handler prefix", ContextPluginName(ctx))
		return nil
	} else if len(this.routers) == 0 {
		this.Printf(ctx, "AddHandlerFuncEx: Not adding handler for %q, no router", ContextPluginName(ctx))
		return nil
	}

	// Set middleware
	for _, name := range ContextHandlerMiddleware(ctx) {
		if middleware, exists := this.middleware[name]; !exists {
			return fmt.Errorf("AddMiddleware %q not found for plugin %q", name, ContextPluginName(ctx))
		} else {
			handler = middleware.AddMiddlewareFunc(ctx, handler)
		}
	}

	// Log adding handler
	if re == nil {
		this.Printf(ctx, "Adding Handler %v => prefix=%q methods=%q middleware=%q", ContextPluginName(ctx), ContextHandlerPrefix(ctx), methods, ContextHandlerMiddleware(ctx))
	} else {
		this.Printf(ctx, "Adding Handler %v => prefix=%q regexp=%q methods=%q middleware=%q", ContextPluginName(ctx), ContextHandlerPrefix(ctx), re, methods, ContextHandlerMiddleware(ctx))
	}

	// Set handler in all routers
	var result error
	for _, router := range this.routers {
		if err := router.AddHandlerFuncEx(ctx, re, handler, methods...); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Return any errors
	return result
}
