package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// Packages
	"github.com/hashicorp/go-multierror"
	"github.com/mutablelogic/go-server/pkg/config"
	context "github.com/mutablelogic/go-server/pkg/context"
)

const (
	flagAddress = "address"
)

func main() {
	// Create flags
	name := filepath.Base(os.Args[0])
	flagset := flag.NewFlagSet(name, flag.ContinueOnError)
	registerArguments(flagset)

	// Parse flags
	if err := flagset.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		} else {
			os.Exit(-1)
		}
	}

	// Get builtin plugins
	plugins, err := BuiltinPlugins()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// TODO: Load dynamic plugins if -plugins flag is set

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
	if plugins, err := config.LoadForResources(fs, resources, plugins); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	} else {
		fmt.Println("plugins=", plugins)
	}

	// Create a context, add flags to context
	ctx := context.ContextForSignal(os.Interrupt)
	ctx = context.WithAddress(ctx, flagset.Lookup(flagAddress).Value.String())

	// Read configuration
	for _, config := range flagset.Args() {
		fmt.Fprintln(os.Stderr, "Reading configuration from", config)
	}

	// Run until done
	fmt.Fprintln(os.Stderr, "Press CTRL+C to exit")
	<-ctx.Done()
	fmt.Fprintln(os.Stderr, "Done")
}

func registerArguments(f *flag.FlagSet) {
	f.String(flagAddress, "", "Override address to listen on")
}
