package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	// If the path is relative, then make it absolute to the binary path
	if !filepath.IsAbs(pluginPath) {
		if strings.HasPrefix(pluginPath, ".") || strings.HasPrefix(pluginPath, "..") {
			if wd, err := os.Getwd(); err == nil {
				pluginPath = filepath.Clean(filepath.Join(wd, pluginPath))
			}
		} else if exec, err := os.Executable(); err == nil {
			pluginPath = filepath.Clean(filepath.Join(filepath.Dir(exec), pluginPath))
		}
	}

	// Load plugins
	plugins, err := provider.LoadPluginsForPattern(pluginPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(plugins)
}
