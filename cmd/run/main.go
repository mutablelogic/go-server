package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// Packages

	"github.com/mutablelogic/go-server/pkg/provider"
)

func main() {
	var pluginPath string
	name := filepath.Base(os.Args[0])
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.StringVar(&pluginPath, "plugin", "*.plugin", "Path to plugins")
	if err := flags.Parse(os.Args[1:]); err != nil {
		os.Exit(1)
	}

	// If the path is relative, then make it absolute to either the binary path
	// or the current working directory
	if !filepath.IsAbs(pluginPath) {
		if strings.HasPrefix(pluginPath, ".") || strings.HasPrefix(pluginPath, "..") {
			if wd, err := os.Getwd(); err == nil {
				pluginPath = filepath.Clean(filepath.Join(wd, pluginPath))
			}
		} else if exec, err := os.Executable(); err == nil {
			pluginPath = filepath.Clean(filepath.Join(filepath.Dir(exec), pluginPath))
		}
	}

	// Create a new provider, load plugins
	provider, err := provider.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := provider.LoadPluginsForPattern(pluginPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Create configurations
	var result error
	for _, plugin := range []string{"logger", "httpserver", "router", "router-frontend", "nginx-handler", "auth-handler", "tokenjar-handler"} {
		// Create a new configuration for the plugin
		if _, err := provider.New(plugin); err != nil {
			result = errors.Join(result, err)
		}
	}
	if result != nil {
		fmt.Fprintln(os.Stderr, result)
		os.Exit(1)
	}

	// Create configuration manually
	// SET does not actually set it:
	// - checks the field exists
	// - updates the dependencies between fields
	// - marks that the value should be set later
	// - there is a binding stage later which actually sets the value
	// - this also ensures we set things in the right order
	// GET will:
	// - check the field exists
	// - return a reference to the value if not applied yet
	provider.Set("logger.flags", []string{"default", "prefix"})

	provider.Set("nginx.binary", "/usr/local/bin/nginx")
	provider.Set("nginx.data", "/var/run/nginx")
	provider.Set("nginx.group", "www-data")

	provider.Set("httpserver.listen", "run/go-server.sock")
	provider.Set("httpserver.group", "www-data")
	provider.Set("httpserver.router", provider.Get("router"))

	provider.Set("auth.tokenjar", provider.Get("tokenjar"))
	provider.Set("auth.tokenbytes", 16)
	provider.Set("auth.bearer", true)

	provider.Set("tokenjar.data", "run")
	provider.Set("tokenjar.writeinterval", "30s")
}

/*

 */
