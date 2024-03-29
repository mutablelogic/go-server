/*
 * go-server - A server for running tasks
 */
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	// Packages
	multierror "github.com/hashicorp/go-multierror"
	config "github.com/mutablelogic/go-server/pkg/config"
	context "github.com/mutablelogic/go-server/pkg/context"
	task "github.com/mutablelogic/go-server/pkg/task"
	version "github.com/mutablelogic/go-server/pkg/version"
)

const (
	flagAddress = "addr"
	flagPlugins = "plugins"
	flagVersion = "version"
)

func main() {
	// Create flags
	flagset := registerArguments(filepath.Base(os.Args[0]))

	// Parse flags
	if err := flagset.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		} else {
			os.Exit(-1)
		}
	}

	// Get builtin plugin prototypes
	protos, err := BuiltinPlugins()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Load dynamic plugins if -plugins flag is set
	if plugins := flagset.Lookup(flagPlugins); plugins != nil && plugins.Value.String() != "" {
		if err := protos.LoadPluginsForPattern(plugins.Value.String(), true); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
	}

	// Check for version flag
	if flagset.Lookup(flagVersion).Value.(flag.Getter).Get().(bool) {
		version.PrintVersion(flagset.Output())
		for key := range protos {
			fmt.Fprintf(flagset.Output(), "  Plugin: %q\n", key)
		}
		os.Exit(0)
	}

	// Read resources using JSON
	var result error
	var resources []config.Resource
	var fs = os.DirFS("/")
	for _, arg := range flagset.Args() {
		// Make path absolute if not already
		if !filepath.IsAbs(arg) {
			if arg_, err := filepath.Abs(arg); err != nil {
				result = multierror.Append(result, err)
				continue
			} else {
				arg = arg_
			}
		}
		if r, err := config.LoadForPattern(fs, strings.TrimPrefix(arg, string(os.PathSeparator))); err != nil {
			result = multierror.Append(result, err)
		} else {
			resources = append(resources, r...)
		}
	}
	if result != nil {
		fmt.Fprintln(os.Stderr, result)
		os.Exit(-1)
	}
	if len(resources) == 0 {
		fmt.Fprintln(os.Stderr, "No resources defined, use -help to see usage")
		os.Exit(-1)
	}

	// Match resources to plugins
	plugins, err := config.LoadForResources(fs, resources, protos)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	} else {
		AddLogPlugin(plugins)
	}

	// Create a context, add flags to context
	ctx := context.ContextForSignal(os.Interrupt, syscall.SIGQUIT)
	ctx = context.WithAddress(ctx, flagset.Lookup(flagAddress).Value.String())

	// Create provider
	provider, err := task.NewProvider(ctx, plugins.Array()...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Receive events from provider
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ch := provider.Sub()
		for event := range ch {
			if err := event.Error(); err != nil {
				provider.Print(event.Context(), err)
			} else {
				provider.Print(event.Context(), event.Key(), ": ", event.Value())
			}
		}
	}()

	// Print out information
	if version.GitSource != "" && version.GitTag != "" {
		fmt.Fprintf(os.Stderr, "%s (%s)\n\n", version.GitSource, version.GitTag)
	}
	for _, key := range provider.Keys() {
		provider.Printf(ctx, "task: %s: %s\n", key, provider.Get(key))
	}

	// Run until done
	fmt.Fprintln(os.Stderr, "\nPress CTRL+C to exit")
	if err := provider.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Wait for goroutine to exit
	wg.Wait()

	// Exit
	fmt.Fprintln(os.Stderr, "Done")
}

func registerArguments(name string) *flag.FlagSet {
	flagset := flag.NewFlagSet(name, flag.ContinueOnError)
	flagset.Usage = func() {
		version.Usage(flagset)
	}
	flagset.String(flagAddress, "", "Override address to listen on")
	flagset.String(flagPlugins, "", "Directory of plugins to load")
	flagset.Bool(flagVersion, false, "Print version and exit")
	return flagset
}
