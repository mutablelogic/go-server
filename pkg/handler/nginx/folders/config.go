/*
Manages the lifecycle of configuration folders for nginx
*/
package folders

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	"github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Config represents the set of files in the available and enabled folders
type Config struct {
	Available *Folder
	Enabled   *Folder
	Ext       string
	Recursive bool
	DirMode   os.FileMode
}

// Template represents a configuration template
type Template struct {
	Name    string // Name should be unique
	Hash    string // Hash of the file contents
	Enabled bool   // Flag indicating if the template is enabled
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create the set of files
func New(available, enabled string, ext string, recursive bool) (*Config, error) {
	f := new(Config)

	// Set basics
	f.DirMode = 0755
	f.Recursive = recursive
	f.Ext = ext

	// Check writable available folder and enumerate files
	if err := isValidDir(available); err != nil {
		return nil, ErrBadParameter.Withf("%v: %q", err, available)
	} else if folder, err := NewFolder(available, ext, recursive); err != nil {
		return nil, err
	} else {
		f.Available = folder
	}

	// Check writable enabled folder and enumerate files
	if err := isValidDir(enabled); err != nil {
		return nil, ErrBadParameter.Withf("%v: %q", err, enabled)
	} else if folder, err := NewFolder(enabled, ext, recursive); err != nil {
		return nil, err
	} else {
		f.Enabled = folder
	}

	// Sync the enabled to available
	if err := f.sync(); err != nil {
		return nil, err
	}

	// Return success
	return f, nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c Config) String() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data)
}

func (t Template) String() string {
	data, _ := json.MarshalIndent(t, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return a list of all available configuration templates
func (c *Config) Templates() []*Template {
	result := make([]*Template, 0, len(c.Available.Files))
	for _, file := range c.Available.Files {
		tmpl := &Template{
			Name: file.Path,
			Hash: file.Hash,
		}
		if _, exists := c.Enabled.Files[file.Hash]; exists {
			tmpl.Enabled = true
		}
		result = append(result, tmpl)
	}
	return result
}

// Return a configuration template by name
func (c *Config) Template(name string) *Template {
	return c.templateByName(name)
}

// Enable a configuration template
func (c *Config) Enable(name string) error {
	if tmpl := c.templateByName(name); tmpl == nil {
		return ErrNotFound.Withf("template: %q", name)
	} else if file, exists := c.Available.Files[tmpl.Hash]; !exists {
		return ErrNotFound.Withf("file: %q", name)
	} else if err := file.Copy(c.Enabled.Root, c.DirMode); err != nil {
		return err
	} else if err := c.Reload(); err != nil {
		return err
	}

	// Return success
	return nil
}

// Disable a configuration template
func (c *Config) Disable(name string) error {
	if tmpl := c.templateByName(name); tmpl == nil {
		return ErrNotFound.Withf("template: %q", name)
	} else if file, exists := c.Enabled.Files[tmpl.Hash]; !exists {
		return ErrNotFound.Withf("file: %q", name)
	} else if err := file.Remove(); err != nil {
		return err
	} else if err := c.Reload(); err != nil {
		return err
	}

	// Return success
	return nil
}

// Reload the configuration
func (c *Config) Reload() error {
	if available, err := NewFolder(c.Available.Root, c.Ext, c.Recursive); err != nil {
		return err
	} else if enabled, err := NewFolder(c.Enabled.Root, c.Ext, c.Recursive); err != nil {
		return err
	} else {
		c.Available = available
		c.Enabled = enabled
	}

	// Sync files
	if err := c.sync(); err != nil {
		return err
	}

	// Return success
	return nil
}

// Create a new configuration template
func (c *Config) Create(name string, body []byte) error {
	// Check for valid name
	if err := c.checkName(name); err != nil {
		return err
	}
	// Check for existing template
	if tmpl := c.templateByName(name); tmpl != nil {
		return ErrDuplicateEntry.Withf("template: %q", name)
	}

	// Create the directory
	abspath := filepath.Join(c.Available.Root, name)
	absdir := filepath.Dir(abspath)
	if err := os.MkdirAll(absdir, c.DirMode); err != nil {
		return err
	}

	// Create the file
	f, err := os.Create(abspath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write the body
	if _, err := f.Write(body); err != nil {
		return err
	}

	// Reload the configuration
	if err := c.Reload(); err != nil {
		return err
	}

	// Return success
	return nil
}

// Delete a configuration template
func (c *Config) Delete(name string) error {
	tmpl := c.templateByName(name)
	if tmpl == nil {
		return ErrNotFound.Withf("template: %q", name)
	}

	// Delete from enabled and available
	var result error
	if file, exists := c.Enabled.Files[tmpl.Hash]; exists {
		result = errors.Join(result, file.Remove())
	}
	if file, exists := c.Available.Files[tmpl.Hash]; exists {
		result = errors.Join(result, file.Remove())
	}

	// Reload the configuration
	result = errors.Join(result, c.Reload())

	// Return any errors
	return result
}

// Render a configuration template
func (c *Config) Render(name string) ([]byte, error) {
	if tmpl := c.templateByName(name); tmpl == nil {
		return nil, ErrNotFound.Withf("template: %q", name)
	} else if file, exists := c.Available.Files[tmpl.Hash]; !exists {
		return nil, ErrNotFound.Withf("file: %q", name)
	} else {
		return file.Render()
	}
}

// Write a new body
func (c *Config) Write(name string, body []byte) error {
	tmpl := c.templateByName(name)
	if tmpl == nil {
		return ErrNotFound.Withf("template: %q", name)
	}

	// Write available
	var result error
	if file, exists := c.Available.Files[tmpl.Hash]; exists {
		if err := file.Write(body); err != nil {
			result = errors.Join(result, err)
		}
	}
	if file, exists := c.Enabled.Files[tmpl.Hash]; exists {
		if err := file.Write(body); err != nil {
			result = errors.Join(result, err)
		}
	}

	// Reload the configuration
	result = errors.Join(result, c.Reload())

	// Return any errors
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (c *Config) checkName(name string) error {
	dir, file := filepath.Split(name)
	if dir != "" && !c.Recursive {
		return ErrBadParameter.Withf("cannot use subdirectories %q", name)
	} else if c.Ext != "" && filepath.Ext(file) != c.Ext {
		return ErrBadParameter.Withf("invalid or missing file extension %q", file)
	} else if !types.IsFilename(file) {
		return ErrBadParameter.Withf("invalid file name %q", name)
	}

	// Return success
	return nil
}

// Return a template by name, or nil
func (c *Config) templateByName(name string) *Template {
	tmpl := c.Templates()
	for _, t := range tmpl {
		if t.Name == name {
			return t
		}
	}
	return nil
}

// Copy any files which are in enabled over to available to ensure that all
// files are available for configuration
func (c *Config) sync() error {
	var result error
	for hash, file := range c.Enabled.Files {
		if _, exists := c.Available.Files[hash]; !exists {
			if err := file.Copy(c.Available.Root, 0755); err != nil {
				result = errors.Join(result, fmt.Errorf("%s: %w", file.Path, err))
			} else if file, err := NewFile(c.Available.Root, file.Path); err != nil {
				result = errors.Join(result, err)
			} else {
				c.Available.Files[file.Hash] = file
			}
		}
	}

	// Return any errors
	return result
}

func isValidDir(dir string) error {
	if stat, err := os.Stat(dir); os.IsNotExist(err) {
		return err
	} else if err != nil {
		return err
	} else if !stat.IsDir() {
		return ErrBadParameter.Withf("not a directory: %q", dir)
	}
	return isWritableFileAtPath(dir)
}
