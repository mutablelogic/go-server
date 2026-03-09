package main

import (
	"fmt"
	"os"

	// Packages
	cmd "github.com/mutablelogic/go-server/pkg/cmd"
	version "github.com/mutablelogic/go-server/pkg/version"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type CLI struct {
	ServerCommands
	ResourcesCommands
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const description = "Kaiak provides an infrastructure as code interface for composing and managing software resources."

var ver = version.Version()

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func main() {
	if err := cmd.Main(CLI{}, description, ver); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
