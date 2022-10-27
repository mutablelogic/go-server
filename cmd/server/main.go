package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mutablelogic/go-server/pkg/context"
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
