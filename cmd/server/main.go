package main

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
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

}

func registerArguments(f *flag.FlagSet) {
	f.String("address", "", "Override address to listen on")
}
