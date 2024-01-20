package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	// Packages
	"github.com/mutablelogic/go-server/pkg/context"
	"github.com/mutablelogic/go-server/pkg/provider"
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

	if provider, err := provider.NewPluginsForPattern(ctx, flags.Plugins); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	} else {
		fmt.Println(provider)
	}
}
