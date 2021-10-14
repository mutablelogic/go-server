package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	// Packages
	config "github.com/mutablelogic/go-server/pkg/config"
	provider "github.com/mutablelogic/go-server/pkg/provider"
	version "github.com/mutablelogic/go-server/pkg/version"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type VarValue struct {
	vars map[string]interface{}
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	flagAddr = "addr"
	flagVar  = "var"
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
	prv, err := provider.NewProvider(ctx, wd, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Report on loaded plugins
	for _, name := range prv.Plugins() {
		prv.Print(provider.ContextWithPluginName(ctx, name), "Loaded: ", prv.GetPlugin(ctx, name))
	}

	// Run the server and plugins
	if err := prv.Run(HandleSignal()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func DefineContext(ctx context.Context, flags *flag.FlagSet) context.Context {
	if addr := flags.Lookup(flagAddr); addr != nil && addr.Value.String() != "" {
		ctx = provider.ContextWithAddr(ctx, addr.Value.String())
	}
	if vars := flags.Lookup(flagVar); vars != nil {
		ctx = provider.ContextWithVars(ctx, vars.Value.(*VarValue).vars)
	}
	return ctx
}

func DefineFlags(args []string) (*flag.FlagSet, error) {
	flags := flag.NewFlagSet(filepath.Base(args[0]), flag.ContinueOnError)

	// Define flags and usage
	flags.String(flagAddr, "", "Override path to unix socket or listening address")
	flags.Var(new(VarValue), flagVar, "Set a variable (name=value)")
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

///////////////////////////////////////////////////////////////////////////////
// VARVALUE IMPLEMENTATION

func (v *VarValue) Set(s string) error {
	if v.vars == nil {
		v.vars = make(map[string]interface{})
	}

	// Split out key-value pair
	kv := strings.SplitN(s, "=", 2)
	switch len(kv) {
	case 1:
		v.vars[kv[0]] = true
	case 2:
		if strings.HasPrefix(kv[1], "\"") && strings.HasSuffix(kv[1], "\"") {
			if q, err := strconv.Unquote(s); err != nil {
				return err
			} else {
				v.vars[kv[0]] = q
			}
		} else {
			v.vars[kv[0]] = kv[1]
		}
	default:
		return ErrBadParameter.Withf("%q", s)
	}

	// Return success
	return nil
}

func (v *VarValue) String() string {
	return ""
}
