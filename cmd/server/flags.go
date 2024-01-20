package main

import (
	"flag"
	"fmt"
	"runtime"

	// Packages
	"github.com/mutablelogic/go-server/pkg/version"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Flags struct {
	Args    []string // Command line arguments
	Plugins string   // Path to plugins

	flagset *flag.FlagSet // Set of parsed flags
	version bool          // When true, print version and exit
	help    bool          // When true, print help and exit
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewFlags(name string, args []string) (*Flags, error) {
	self := &Flags{
		flagset: flag.NewFlagSet(name, flag.ContinueOnError),
	}
	self.flagset.Usage = func() {
		self.PrintUsage()
	}

	// Register the flags
	self.RegisterFlags()

	// Parse arguments
	if err := self.flagset.Parse(args); err != nil {
		return nil, err
	} else {
		self.Args = self.flagset.Args()
	}

	// At the moment, no arguments are allowed
	if len(self.Args) != 0 {
		return nil, ErrBadParameter.With("Presently no arguments are allowed")
	}

	// Check for version
	if self.version {
		self.PrintVersion()
		return self, flag.ErrHelp
	} else if self.help {
		if self.Plugins == "" {
			self.flagset.Usage()
		}
		return self, flag.ErrHelp
	}

	// Return success
	return self, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

/*
// Prints usage of a plugin
// TODO: Use blocks for sub-blocks (ie, tls {} instead of tls.key, etc)
func (self *Flags) PrintPluginUsage(path string, meta *hcl.BlockMeta) {
	w := self.flagset.Output()
	if meta.Description != "" {
		fmt.Fprintf(w, "// %s: %s\n", meta.Name, meta.Description)
	} else {
		fmt.Fprintf(w, "// %s\n", meta.Name)
	}
	fmt.Fprintf(w, "// plugin: %s\n", path)
	if meta.Label != nil {
		fmt.Fprintf(w, "%s \"<label>\" {\n", meta.Name)
	} else {
		fmt.Fprintf(w, "%s {\n", meta.Name)
	}
	if len(meta.Attr) > 0 {
		for i, attr := range meta.Attr {
			if i > 0 {
				fmt.Fprintln(w, "")
			}
			if attr.Description != "" {
				fmt.Fprintf(w, "  // %s\n", attr.Description)
			}
			if attr.Required {
				fmt.Fprint(w, "  // Note: required\n")
			}
			fmt.Fprintf(w, "  %s = <%s>\n", attr.Label, attr.Type)
		}
	}
	fmt.Fprintln(w, "}\n")
}
*/

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (self *Flags) RegisterFlags() {
	self.flagset.BoolVar(&self.version, "version", false, "Print version and exit")
	self.flagset.BoolVar(&self.help, "help", false, "Print help and exit")
	self.flagset.StringVar(&self.Plugins, "plugins", "", "Path to plugins")
}

func (self *Flags) PrintUsage() {
	name := self.flagset.Name()
	w := self.flagset.Output()
	fmt.Fprintf(w, "Usage of %s:\n", name)
	fmt.Fprintf(w, "  %s -version\n", name)
	fmt.Fprintln(w, "  \tPrint version information and exit")
	fmt.Fprintf(w, "  %s -help\n", name)
	fmt.Fprintln(w, "  \tPrint this help information and exit")
	fmt.Fprintf(w, "  %s -plugins <plugins> -help \n", name)
	fmt.Fprintln(w, "  \tProvide information on plugins and configuration and exit")
	fmt.Fprintf(w, "  %s -plugins <plugins>\n", name)
	fmt.Fprintln(w, "  \tRun the server with the plugins until interrupted or signal caught")
	fmt.Fprintln(w, "\nFlags:")
	self.flagset.PrintDefaults()
}

func (self *Flags) PrintVersion() {
	w := self.flagset.Output()
	if version.GitSource != "" {
		fmt.Fprintf(w, "  Url: https://%v\n", version.GitSource)
	}
	if version.GitTag != "" || version.GitBranch != "" {
		fmt.Fprintf(w, "  Version: %v (branch: %q hash:%q)\n", version.GitTag, version.GitBranch, version.GitHash)
	}
	if version.GoBuildTime != "" {
		fmt.Fprintf(w, "  Built: %v\n", version.GoBuildTime)
	}
	fmt.Fprintf(w, "  Go: %v (%v/%v)\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
