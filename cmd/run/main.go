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
	for _, plugin := range []string{"logger", "httpserver", "router", "nginx-handler", "auth-handler", "tokenjar-handler"} {
		if _, err := provider.New(plugin); err != nil {
			result = errors.Join(result, err)
		}
	}
	if result != nil {
		fmt.Fprintln(os.Stderr, result)
		os.Exit(1)
	}

	// TODO: Set parameters from a JSON file
	provider.Set("logger.flags", []string{"default", "prefix"})
}

/*
{
	"logger": {
		"flags": ["default", "prefix"]
	},
	"nginx": {
		"binary": "/usr/local/bin/nginx",
		"data": "/var/run/nginx",
		"group": "www-data",
	},
	httpserver": {
		"listen": "run/go-server.sock",
		"group": "www-data",
		"router": "${ router }",
	},
	"router": {
		"services": {
			"nginx": {
				"service": "${ nginx }",
				"middleware": ["logger", "auth"]
			},
			"auth": {
				"service": "${ auth }",
				"middleware": ["logger", "auth"]
			},
			"router": {
				"service": "${ router }",
				"middleware": ["logger", "auth"]
			},
	},
	"auth": {
		"tokenjar": "${ tokenjar }",
		"tokenbytes": 16,
		"bearer": true,
	},
	"tokenjar": {
		"data": "run",
		"writeinterval": "30s",
	},
}
*/
