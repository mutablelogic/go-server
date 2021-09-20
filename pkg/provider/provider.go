package provider

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"plugin"
	"strconv"
	"sync"
	"syscall"

	// Packages
	marshaler "github.com/djthorpe/go-marshaler"
	config "github.com/djthorpe/go-server/pkg/config"
	multierror "github.com/hashicorp/go-multierror"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type provider struct {
	plugins map[string]*plugincfg

	// List of interfaces
	names      []string
	loggers    []Logger
	routers    []Router
	middleware map[string]Middleware
	queue      []EventQueue
}

type plugincfg struct {
	path    string
	plugin  Plugin
	handler config.Handler
	config  map[string]interface{}
}

type PluginUsageFunc func(io.Writer)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	PluginFileExt = ".plugin"
)

const (
	funcName  = "Name"
	funcUsage = "Usage"
	funcNew   = "New"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewProvider(parent context.Context, basepath string, cfg *config.Config) (*provider, error) {
	this := new(provider)
	this.plugins = make(map[string]*plugincfg, len(cfg.Plugin))
	this.middleware = make(map[string]Middleware, len(cfg.Plugin))

	// Return error if no plugins
	if len(cfg.Plugin) == 0 {
		return nil, ErrBadParameter.With("No plugins defined")
	}

	// Read plugin names
	var result error
	for _, path := range cfg.Plugin {
		if path := PluginPath(basepath, path); path == "" {
			result = multierror.Append(result, ErrNotFound.With("Plugin: ", path))
		} else if stat, err := os.Stat(path); err != nil {
			result = multierror.Append(result, ErrNotFound.With("Plugin: ", path))
		} else if !stat.Mode().IsRegular() {
			result = multierror.Append(result, ErrBadParameter.With("Plugin: ", path))
		} else if name, err := GetPluginName(path); err != nil {
			result = multierror.Append(result, ErrBadParameter.With("Plugin: ", err))
		} else if _, exists := this.plugins[name]; exists {
			result = multierror.Append(result, ErrDuplicateEntry.With("Plugin: ", name))
		} else {
			this.names = append(this.names, name)
			this.plugins[name] = &plugincfg{path, nil, cfg.Handler[name], nil}
		}
	}

	// Warn when a handler is not associated with a plugin name
	for name, handler := range cfg.Handler {
		if this.plugins[name] == nil {
			result = multierror.Append(result, ErrNotFound.With("Plugin: ", name))
		}
		if handler.Prefix == "" {
			result = multierror.Append(result, ErrBadParameter.With("Missing prefix for handler: ", name))
		}
		for _, middleware := range handler.Middleware {
			if this.plugins[middleware] == nil && name != middleware {
				result = multierror.Append(result, ErrNotFound.With("Missing middleware "+strconv.Quote(middleware)+" for handler: ", name))
			}
		}
	}

	// Obtain plugin configurations
	for _, name := range this.names {
		if cfg, exists := cfg.Config[name]; exists {
			if plugincfg, exists := this.plugins[name]; exists {
				if cfg, ok := cfg.(map[string]interface{}); ok {
					plugincfg.config = cfg
				} else {
					result = multierror.Append(result, ErrBadParameter.With("Invalid configuration for: ", name))
				}
			}
		}
	}

	// Return any errors
	if result != nil {
		return nil, result
	}

	// Load all plugins
	for _, name := range this.names {
		// If already loaded then skip
		plugincfg := this.plugins[name]
		if plugincfg.plugin != nil {
			continue
		}

		// Set plugin context, load plugin
		ctx := ContextWithPluginName(parent, name)
		if plugincfg.handler.Prefix != "" {
			ctx = ContextWithHandler(ctx, plugincfg.handler)
		}
		plugin := this.GetPlugin(ctx, name)
		if plugin == nil {
			result = multierror.Append(result, ErrInternalAppError.With(name))
			continue
		}
	}

	// Return errors from initialisation
	if result != nil {
		return nil, result
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *plugincfg) String() string {
	str := "<plugin"
	if this.path != "" {
		str += fmt.Sprintf(" path=%q", this.path)
	}
	if this.plugin != nil {
		str += fmt.Sprintf(" plugin=%v", this.plugin)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *provider) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	var cancels []context.CancelFunc
	var result error

	// Receive any errors
	/*
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case err := <-errs:
					result = multierror.Append(result, err)

					// Send terminate signal to process

				}
			}
		}()
	*/

	// Run all plugins and wait until done
	this.Print(ctx, "Running plugins:")
	for name, cfg := range this.plugins {
		wg.Add(1)
		ctx, cancel := context.WithCancel(context.Background())
		cancels = append(cancels, cancel)
		go func(name string, cfg *plugincfg) {
			defer wg.Done()
			this.Print(ctx, " ", name, " running")
			if err := cfg.plugin.Run(ContextWithHandler(ContextWithPluginName(ctx, name), cfg.handler), this); err != nil {
				this.Print(ctx, " ", name, " error: ", err)
				result = multierror.Append(result, err)
				if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
					result = multierror.Append(result, err)
				}
			}
			this.Print(ctx, " ", name, " stopped")
		}(name, cfg)
	}

	// Wait for cancel
	<-ctx.Done()

	// Cancel all plugins
	this.Print(ctx, "Stopping plugins:")
	for _, cancel := range cancels {
		cancel()
	}

	// Wait for all plugins to finish
	wg.Wait()

	// Return any errors
	return result
}

func (this *provider) Plugins() []string {
	result := make([]string, 0, len(this.names))
	for _, name := range this.names {
		result = append(result, name)
	}
	return result
}

func (this *provider) GetPlugin(ctx context.Context, name string) Plugin {
	plugincfg, exists := this.plugins[name]
	if exists && plugincfg.plugin != nil {
		return plugincfg.plugin
	} else if !exists {
		this.Print(ctx, "GetPlugin: ", ErrNotFound.With(name))
		return nil
	}
	plugin, err := this.pluginWithPath(ctx, name, plugincfg.path)
	if err != nil {
		this.Print(ctx, "GetPlugin: ", err)
		return nil
	} else if err := this.setPlugin(ctx, name, plugin); err != nil {
		this.Print(ctx, "GetPlugin: ", err)
		return nil
	}
	// Return success
	return plugin
}

func (this *provider) setPlugin(ctx context.Context, name string, plugin Plugin) error {
	plugincfg := this.plugins[name]
	plugincfg.plugin = plugin
	if router, ok := plugin.(Router); ok {
		this.routers = append(this.routers, router)
	}
	if logger, ok := plugin.(Logger); ok {
		this.loggers = append(this.loggers, logger)
	}
	if middleware, ok := plugin.(Middleware); ok {
		this.middleware[name] = middleware
	}
	if queue, ok := plugin.(EventQueue); ok {
		this.queue = append(this.queue, queue)
	}
	return nil
}

func (this *provider) GetConfig(ctx context.Context, v interface{}) error {
	name := ContextPluginName(ctx)
	if name == "" {
		return ErrBadParameter.With("Missing plugin name")
	} else if plugincfg, exists := this.plugins[name]; !exists {
		return ErrNotFound.With("GetConfig: ", name)
	} else if plugincfg.config == nil {
		// No configuration for this plugin
		return nil
	} else {
		return marshaler.NewDecoder("yaml", marshaler.ConvertIntUint, marshaler.ConvertMapInterface, marshaler.ConvertDuration, marshaler.ConvertTime).Decode(plugincfg.config, v)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// PluginPath returns absolute path to a plugin or empty string
// if it can't be located
func PluginPath(basepath, path string) string {
	if filepath.IsAbs(path) {
		return path
	} else if abs, err := filepath.Abs(filepath.Join(basepath, path)); err != nil {
		return ""
	} else if ext := filepath.Ext(path); ext == "" {
		return ext + PluginFileExt
	} else {
		return abs
	}
}

// GetPluginName returns the plugin name from a file path
func GetPluginName(path string) (string, error) {
	plugin, err := plugin.Open(path)
	if err != nil {
		return "", err
	}
	// Return module name
	if fn, err := plugin.Lookup(funcName); err != nil {
		return "", err
	} else if name := fn.(func() string)(); name == "" {
		return "", ErrInternalAppError.With("Name returned nil: ", path)
	} else {
		return name, nil
	}
}

// GetPluginUsage returns the usage function for a plugin
func GetPluginUsage(path string) (PluginUsageFunc, error) {
	if plugin, err := plugin.Open(path); err != nil {
		return nil, err
	} else if fn, err := plugin.Lookup(funcUsage); err != nil {
		return nil, err
	} else {
		return fn.(func(io.Writer)), nil
	}
}

// pluginWithPath returns a plugin from a file path
func (this *provider) pluginWithPath(ctx context.Context, name, path string) (Plugin, error) {
	// If plugin is already in the path, return an error, as it would be a circular reference
	if ContextHasPluginParent(ctx, name) {
		return nil, ErrDuplicateEntry.With(name)
	} else {
		ctx = ContextWithPluginName(ctx, name)
	}

	// If plugin already exists, return it
	plugincfg, exists := this.plugins[name]
	if exists && plugincfg.plugin != nil {
		return plugincfg.plugin, nil
	}

	// Create a new module from plugin
	if plugin, err := plugin.Open(plugincfg.path); err != nil {
		return nil, err
	} else if fn, err := plugin.Lookup(funcNew); err != nil {
		return nil, err
	} else if fn_, ok := fn.(func(context.Context, Provider) Plugin); !ok {
		return nil, ErrInternalAppError.With("New returned nil: ", name)
	} else {
		if module := fn_(ctx, this); module == nil {
			return nil, ErrInternalAppError.With("New returned nil: ", name)
		} else {
			plugincfg.plugin = module
		}
	}

	// Return success
	return plugincfg.plugin, nil
}
