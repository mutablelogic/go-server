package config

import (
	"flag"
	"fmt"
	"io"
	"runtime"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	GitSource   string
	GitTag      string
	GitBranch   string
	GitHash     string
	GoBuildTime string
)

func PrintVersion(w io.Writer) {
	if GitSource != "" {
		fmt.Fprintf(w, "  URL: https://%v\n", GitSource)
	}
	if GitTag != "" || GitBranch != "" {
		fmt.Fprintf(w, "  Version: %v (branch: %q hash:%q)\n", GitTag, GitBranch, GitHash)
	}
	if GoBuildTime != "" {
		fmt.Fprintf(w, "  Build Time: %v\n", GoBuildTime)
	}
	fmt.Fprintf(w, "  Go: %v (%v/%v)\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Usage(w io.Writer, flags *flag.FlagSet) {
	name := flags.Name()

	flags.SetOutput(w)

	fmt.Fprintf(w, "%s: monolith server\n", name)
	fmt.Fprintf(w, "\nUsage:\n")
	fmt.Fprintf(w, "  %s <flags> config.yaml\n", name)
	fmt.Fprintf(w, "  %s -help\n", name)
	fmt.Fprintf(w, "  %s -help <plugin>\n", name)

	fmt.Fprintln(w, "\nFlags:")
	flags.PrintDefaults()

	fmt.Fprintln(w, "\nVersion:")
	PrintVersion(w)
	fmt.Fprintln(w, "")
}
