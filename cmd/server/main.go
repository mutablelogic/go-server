package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	// Modules
	config "github.com/djthorpe/go-server/pkg/config"
	provider "github.com/djthorpe/go-server/pkg/provider"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	flagAddr      = flag.String("addr", "", "Path to unix socket or listening address")
	flagNamespace = flag.String("ns", "", "Namespace for configuration options")
)

///////////////////////////////////////////////////////////////////////////////
// MAIN

func main() {
	flag.Usage = func() {
		config.Usage(os.Stdout)
	}
	flag.Parse()

	// Get configuration filenames
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Syntax error: Expected configuration file argument")
		os.Exit(-1)
	}

	// Read configuration file
	cfg, err := config.New(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Create a provider with the configuration
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
	provider, err := provider.NewProvider(wd, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	ctx := HandleSignal()
	provider.Print(ctx, "Loaded plugins:")
	for _, name := range provider.Plugins() {
		provider.Print(ctx, " ", name, " => ", provider.GetPlugin(ctx, name))
	}

	// Run the server
	if err := provider.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func HandleSignal() context.Context {
	// Handle signals - call cancel when interrupt received
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		cancel()
	}()
	return ctx
}
