package hclconfig

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	// Namespace imports
	. "github.com/djthorpe/go-errors"

	// Modules
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	// Map plugin name to plugin description
	Plugins map[string]*Plugin
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	fileExtPlugin = ".plugin"
	fileExtHCL    = ".hcl"
	fileExtJSON   = ".json"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates an empty configuration from a working directory and path to plugins.
// Uses the plugins loaded to create a "specification" for reading the configuration
// file.
func New(root, path string) (*Config, error) {
	this := new(Config)
	this.Plugins = make(map[string]*Plugin)

	// Load plugins
	if err := fs.WalkDir(os.DirFS("/"), abspath(root, path), func(path string, info fs.DirEntry, err error) error {
		return this.walkplugins(path, info, err)
	}); err != nil {
		return nil, err
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c *Config) String() string {
	str := "<config"
	for _, plugin := range c.Plugins {
		str += fmt.Sprint(" ", plugin)
	}
	return str + ">"
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Config) Parse(root, path string) error {
	parser := hclparse.NewParser()

	// Load HCL or JSON files
	filesys := os.DirFS("/")
	if err := fs.WalkDir(filesys, abspath(root, path), func(path string, info fs.DirEntry, err error) error {
		return c.walkconfig(parser, filesys, path, info, err)
	}); err != nil {
		return err
	}

	// Merge files together
	files := parser.Files()
	body := make([]*hcl.File, 0, len(files))
	for _, file := range files {
		body = append(body, file)
	}

	// Create a specification for parsing
	tuples := make([]string, 0, len(c.Plugins))
	spec := make(hcldec.TupleSpec, 0, len(c.Plugins))
	for _, plugin := range c.Plugins {
		tuples = append(tuples, plugin.Name)
		spec = append(spec, plugin.Spec)
	}

	// Parse the configuration into a cty.Value
	value, diags := hcldec.Decode(hcl.MergeFiles(body), spec, &hcl.EvalContext{
		Variables: map[string]cty.Value{},
	})
	if diags.HasErrors() {
		return diags
	}

	// Convert cty.Value into configuration objects for plugins
	var result error
	value.ForEachElement(func(key, tuple cty.Value) bool {
		var i int
		if err := gocty.FromCtyValue(key, &i); err != nil {
			result = multierror.Append(result, ErrInternalAppError)
			return true
		}

		// Get the plugin
		plugin, exists := c.Plugins[tuples[i]]
		if !exists {
			result = multierror.Append(result, ErrInternalAppError)
			return true
		}

		// Append plugin resources
		return tuple.ForEachElement(func(_, resource cty.Value) bool {
			if err := plugin.Append(resource); err != nil {
				result = multierror.Append(result, err)
				return true
			} else {
				return false
			}
		})
	})

	// Return success
	return result
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (c *Config) walkplugins(path string, d fs.DirEntry, err error) error {
	// Pass down any error
	if err != nil {
		return err
	}

	// Ignore any hidden files
	if strings.HasPrefix(d.Name(), ".") {
		if d.IsDir() {
			return fs.SkipDir
		} else {
			return nil
		}
	}

	// Recurse into folders
	if d.IsDir() {
		return nil
	}

	// Return error if not a regular file
	if !d.Type().IsRegular() {
		return ErrNotImplemented.Withf("%q", d.Name())
	}

	// Only process .plugin files
	switch strings.ToLower(filepath.Ext(d.Name())) {
	case fileExtPlugin:
		if plugin, err := c.OpenPlugin(string(os.PathSeparator) + path); err != nil {
			return err
		} else if _, exists := c.Plugins[plugin.Name]; exists {
			return ErrDuplicateEntry.With(plugin.Name)
		} else {
			c.Plugins[plugin.Name] = plugin
		}
	}

	return nil
}

func (c *Config) walkconfig(parser *hclparse.Parser, filesys fs.FS, path string, d fs.DirEntry, err error) error {
	// Pass down any error
	if err != nil {
		return err
	}
	// Ignore any hidden files
	if strings.HasPrefix(d.Name(), ".") {
		if d.IsDir() {
			return fs.SkipDir
		} else {
			return nil
		}
	}
	// Recurse into folders
	if d.IsDir() {
		return nil
	}
	// Return error if not a regular file
	if !d.Type().IsRegular() {
		return ErrNotImplemented.Withf("%q", d.Name())
	}
	// Deal with filetypes
	switch strings.ToLower(filepath.Ext(d.Name())) {
	case fileExtHCL:
		if data, err := readall(filesys, path); err != nil {
			return err
		} else if _, diags := parser.ParseHCL(data, path); diags.HasErrors() {
			return diags
		}
	case fileExtJSON:
		if data, err := readall(filesys, path); err != nil {
			return err
		} else if _, diags := parser.ParseJSON(data, path); diags.HasErrors() {
			return diags
		}
	default:
		return ErrNotImplemented.Withf("%q", d.Name())
	}
	return nil
}

// Return an absolute path, but without the first '/'
func abspath(root, path string) string {
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	return strings.TrimPrefix(path, string(os.PathSeparator))
}

// Read all bytes from regular file
func readall(filesys fs.FS, path string) ([]byte, error) {
	if data, err := fs.ReadFile(filesys, path); err != nil {
		return nil, fmt.Errorf("%q: %w", filepath.Base(path), err)
	} else {
		return data, nil
	}
}
