package main

import (
	"flag"
	"os"

	// Packages
	nginx "github.com/mutablelogic/go-server/pkg/nginx/client"
	version "github.com/mutablelogic/go-server/pkg/version"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Flags represents the command line flags
type Flags struct {
	*flag.FlagSet
	cmd []Command
}

type Command struct {
	Name  string
	Usage string
	Fn    CommandFunc
}

type CommandFunc func(*Flags, *nginx.Client) error

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	flagEndpoint = "endpoint"
	flagToken    = "token"
	flagVersion  = "version"
	flagDebug    = "debug"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewFlags(name string) *Flags {
	flags := &Flags{
		FlagSet: flag.NewFlagSet(name, flag.ContinueOnError),
	}
	flags.Usage = func() {
		// TODO
		version.Usage(flags.FlagSet)
	}

	// Register flags
	flags.String(flagEndpoint, "${GOSERVER_ENDPOINT}", "Endpoint for the server")
	flags.String(flagToken, "${GOSERVER_TOKEN}", "Authorization token")
	flags.Bool(flagVersion, false, "Print version and exit")
	flags.Bool(flagDebug, false, "Verbose output on requests and responses")

	// Register commands
	flags.cmd = []Command{
		{
			Name:  "version",
			Usage: "Print version and exit",
			Fn:    Version,
		}, {
			Name:  "test",
			Usage: "Test nginx configuration",
			Fn:    Test,
		},
	}

	return flags
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (f *Flags) GetVersion() bool {
	return f.Lookup(flagVersion).Value.(flag.Getter).Get().(bool)
}

func (f *Flags) GetDebug() bool {
	return f.Lookup(flagDebug).Value.(flag.Getter).Get().(bool)
}

func (f *Flags) GetEndpoint() string {
	return os.ExpandEnv(f.Lookup(flagEndpoint).Value.(flag.Getter).String())
}

func (f *Flags) GetToken() string {
	return os.ExpandEnv(f.Lookup(flagToken).Value.(flag.Getter).String())
}

func (f *Flags) GetCommand() (CommandFunc, error) {
	if f.NArg() == 0 {
		return nil, ErrBadParameter.With("no command specified, use -help for usage")
	}
	for _, cmd := range f.cmd {
		if cmd.Name == f.Arg(0) {
			return cmd.Fn, nil
		}
	}
	return nil, ErrNotFound.Withf("command: %q", f.Arg(0))
}

func (f *Flags) PrintVersion() {
	version.PrintVersion(f.Output())
}
