package config

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	// Modules
	yaml "gopkg.in/yaml.v3"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Plugin  []string           `yaml:"plugins"`
	Handler map[string]Handler `yaml:"handlers"`
	Config  map[string]interface{}
}

type Handler struct {
	Prefix     string   `yaml:"prefix"`
	Middleware []string `yaml:"middleware"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(pattern string) (*Config, error) {
	this := new(Config)

	// Glob files, sort alphabetically
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	sort.Strings(files)

	// yaml decode plugins and handlers
	for _, path := range files {
		// Read configuration file
		r, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer r.Close()
		dec := yaml.NewDecoder(r)
		if err := dec.Decode(&this); err != nil {
			return nil, err
		}
	}

	// In the second pass, read plugin configurations
	for _, path := range files {
		// Read configuration file
		r, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer r.Close()
		dec := yaml.NewDecoder(r)
		if err := dec.Decode(&this.Config); err != nil {
			return nil, err
		}
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c *Config) String() string {
	str := "<config"
	if len(c.Plugin) > 0 {
		str += fmt.Sprintf(" plugins=%q", c.Plugin)
	}
	return str + ">"
}

func (h Handler) String() string {
	str := "<handler"
	if h.Prefix != "" {
		str += fmt.Sprintf(" prefix=%q", h.Prefix)
	}
	if len(h.Middleware) > 0 {
		str += fmt.Sprintf(" middleware=%q", h.Middleware)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// USAGE

func Usage(w io.Writer, flags *flag.FlagSet) {
	name := flags.Name()

	flags.SetOutput(w)

	fmt.Fprintf(w, "%s: Monolith server\n", name)
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
