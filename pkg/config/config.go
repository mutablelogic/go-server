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

	// Namespace imports
	. "github.com/djthorpe/go-errors"
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

func New(patterns ...string) (*Config, error) {
	this := new(Config)
	this.Handler = make(map[string]Handler)

	// Glob files, sort alphabetically
	var files []string
	for _, pattern := range patterns {
		files_, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		} else if len(files_) == 0 {
			return nil, ErrNotFound.Withf("%q", pattern)
		}
		files = append(files, files_...)
	}
	sort.Strings(files)

	// yaml decode plugins and handlers and merge into a single
	for _, path := range files {
		// Read configuration file
		r, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer r.Close()

		// Decode into a configuration
		var config Config
		if err := yaml.NewDecoder(r).Decode(&config); err != nil {
			return nil, fmt.Errorf("%q: %w", path, err)
		}

		// Merge in plugins, ignore duplicates
		for _, plugin := range config.Plugin {
			if !sliceContains(this.Plugin, plugin) {
				this.Plugin = append(this.Plugin, plugin)
			}
		}

		// Marge in handlers, error when there are duplicates
		for key, handler := range config.Handler {
			if _, exists := this.Handler[key]; exists {
				return nil, ErrDuplicateEntry.Withf("Handler: %q", key)
			} else {
				this.Handler[key] = handler
			}
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

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func sliceContains(arr []string, elem string) bool {
	for _, v := range arr {
		if v == elem {
			return true
		}
	}
	return false
}
