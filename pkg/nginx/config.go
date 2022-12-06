package nginx

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	sync.RWMutex

	// Configuration Files
	Files map[string]*File

	// Templates
	t *template.Template
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewConfig(available, enabled string) (*Config, error) {
	this := new(Config)
	this.Files = make(map[string]*File, 10)
	tmpl := make([]string, 0, 10)

	// Get available files
	if available != "" {
		files, err := enumerate(available, true)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			key := file.EnabledBase()
			if _, exists := this.Files[key]; !exists {
				this.Files[key] = file
				tmpl = append(tmpl, filepath.Join(available, file.Path))
			} else {
				return nil, ErrDuplicateEntry.Withf("Duplicate file: %q", file.Path)
			}
		}
	}

	// Compile templates for available files, and return any errors
	if t, err := template.ParseFiles(tmpl...); err != nil {
		return nil, err
	} else {
		this.t = t
	}

	// Set enabled flag on files
	if enabled != "" {
		files, err := enumerate(available, true)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			key := file.EnabledBase()
			if file, exists := this.Files[key]; exists {
				file.Enabled = true
			}
		}
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// Return a list of files which are available to be enabled
func (c *Config) String() string {
	c.RLock()
	defer c.RUnlock()

	str := "<nginx.config"
	for _, file := range c.Files {
		str += fmt.Sprint(" ", file)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return a list of files which are available to be enabled
func (c *Config) Available() []string {
	c.RLock()
	defer c.RUnlock()

	result := make([]string, 0, len(c.Files))
	for key := range c.Files {
		result = append(result, key)
	}
	return result
}

func (c *Config) Enable(key string, args ...any) error {
	c.Lock()
	defer c.Unlock()

	_, exists := c.Files[key]
	// Check to make sure configuration is available
	if !exists {
		return ErrNotFound.With(key)
	}

	// TODO
	return ErrNotImplemented
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Enumerates current set of files within a folder
func enumerate(root string, recursive bool) ([]*File, error) {
	var result []*File

	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		// Skip errors
		if err != nil {
			return err
		}

		// Ignore hidden files
		if strings.HasPrefix(d.Name(), ".") && path != root {
			if d.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}

		// Recurse into directories
		if d.IsDir() && path != root {
			if recursive {
				return nil
			} else {
				return filepath.SkipDir
			}
		}

		// Only enumerate regular files, ignore any files with errors
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if !validFileMode(info.Mode()) {
			return nil
		}
		if relpath, err := filepath.Rel(root, path); err == nil {
			result = append(result, &File{Path: relpath, Info: info})
		}

		// Return success
		return nil
	}); err != nil {
		return nil, err
	}

	// Return success
	return result, nil
}

func validFileMode(mode fs.FileMode) bool {
	if mode.IsRegular() {
		return true
	}
	if mode.Type() == fs.ModeSymlink {
		return true
	}
	return false
}
