package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	// Modules
	config "github.com/mutablelogic/go-server/pkg/config"
	provider "github.com/mutablelogic/go-server/pkg/provider"
	version "github.com/mutablelogic/go-server/pkg/version"
)

const (
	flagAddr = "addr"
)

///////////////////////////////////////////////////////////////////////////////
// MAIN

func main() {
	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Define flags
	flags, err := DefineFlags(os.Args)
	if err == flag.ErrHelp {
		for _, arg := range flags.Args() {
			if path := provider.PluginPath(wd, arg); path == "" {
				continue
			} else if name, err := provider.GetPluginName(path); err == nil && name != "" {
				fmt.Fprintf(flags.Output(), "Usage of plugin %q:\n", name)
				if fn, err := provider.GetPluginUsage(path); err != nil {
					fmt.Fprintln(flags.Output(), " ", err)
				} else if fn != nil {
					fn(flags.Output())
				}
				fmt.Fprintln(flags.Output(), "")
			}
		}
		os.Exit(0)
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Context with flags
	ctx := DefineContext(context.Background(), flags)

	// Read configuration files
	cfg, err := config.New(flags.Args()...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Create a provider with the configuration
	provider, err := provider.NewProvider(ctx, wd, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
	provider.Print(ctx, "Config: ", cfg)

	// Report on loaded plugins
	provider.Print(ctx, "Loaded plugins:")
	for _, name := range provider.Plugins() {
		provider.Print(ctx, " ", name, " => ", provider.GetPlugin(ctx, name))
	}

	// Run the server and plugins
	if err := provider.Run(HandleSignal()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func DefineContext(ctx context.Context, flags *flag.FlagSet) context.Context {
	if addr := flags.Lookup(flagAddr); addr != nil {
		ctx = provider.ContextWithAddr(ctx, addr.Value.String())
	}
	return ctx
}

func DefineFlags(args []string) (*flag.FlagSet, error) {
	flags := flag.NewFlagSet(filepath.Base(args[0]), flag.ContinueOnError)

	// Define flags and usage
	flags.String(flagAddr, "", "Override path to unix socket or listening address")
	flags.Usage = func() {
		version.Usage(os.Stdout, flags)
	}

	// Parse flags from comand line, return on error
	if err := flags.Parse(args[1:]); err != nil {
		return flags, err
	}

	// Check for arguments
	if flags.NArg() < 1 {
		return nil, fmt.Errorf("syntax error: expected configuration file argument")
	}

	// Return success
	return flags, nil
}

func HandleSignal() context.Context {
	// Handle signals - call cancel when interrupt received
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		cancel()
	}()
	return ctx
}
