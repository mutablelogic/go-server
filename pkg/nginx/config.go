package nginx

import (
	"io/fs"
	"path/filepath"
	"strings"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	// Configuration Files
	Files map[string]*File
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewConfig(available, enabled string) (*Config, error) {
	this := new(Config)
	this.Files = make(map[string]*File, 10)

	// Get available files
	if available != "" {
		files, err := enumerate(available, true)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			key := file.AvailableBase()
			if _, exists := this.Files[key]; !exists {
				this.Files[key] = file
			} else {
				return nil, ErrDuplicateEntry.Withf("Duplicate file: %q", file.Path)
			}
		}
	}

	// Get enabled files
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
		// TODO
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Config) Available() []string {
	result := make([]string, 0, len(c.Files))
	for _, file := range c.Files {
		result = append(result, file.AvailableBase())
	}
	return result
}

func (c *Config) Enable(key string) error {
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
