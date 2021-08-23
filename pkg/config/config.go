package config

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

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

func New(path string) (*Config, error) {
	this := new(Config)

	// Read configuration file
	if r, err := os.Open(path); err != nil {
		return nil, err
	} else {
		defer r.Close()
		dec := yaml.NewDecoder(r)
		if err := dec.Decode(&this); err != nil {
			return nil, err
		}
	}

	// In the second pass, read plugin configurations
	if r, err := os.Open(path); err != nil {
		return nil, err
	} else {
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

func Usage(w io.Writer) {
	name := filepath.Base(flag.CommandLine.Name())
	fmt.Fprintf(w, "%s: Monolith server\n", name)
	fmt.Fprintf(w, "\nUsage:\n")
	fmt.Fprintf(w, "  %s <flags> config.yaml\n", name)

	fmt.Fprintln(w, "\nFlags:")
	flag.PrintDefaults()

	fmt.Fprintln(w, "\nVersion:")
	PrintVersion(w)
	fmt.Fprintln(w, "")
}
