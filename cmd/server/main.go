package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	// Packages
	hcl "github.com/mutablelogic/go-hcl/pkg/block"
	context "github.com/mutablelogic/go-server/pkg/context"
	provider "github.com/mutablelogic/go-server/pkg/provider"
)

func main() {
	// Create flags and context
	flags, err := NewFlags(filepath.Base(os.Args[0]), os.Args[1:])
	ctx := context.ContextForSignal(os.Interrupt, syscall.SIGQUIT)

	// If any error returned, then print error and exit
	if errors.Is(err, flag.ErrHelp) {
		os.Exit(0)
	} else if err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Load plugins into the provider
	if flags.Plugins == "" {
		fmt.Fprintln(os.Stderr, "Missing -plugins argument")
		os.Exit(-1)
	}
	server, err := provider.NewPluginsForPattern(ctx, flags.Plugins)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Create configuration
	if _, err := server.NewBlock("logger", nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	} else if httpserver, err := server.NewBlock("httpserver", hcl.NewLabel("main")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	} else if simplerouter, err := server.NewBlock("simplerouter", hcl.NewLabel("main")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	} else if err := server.Set(httpserver, hcl.NewLabel("router"), simplerouter); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Run the application
	if err := server.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}
