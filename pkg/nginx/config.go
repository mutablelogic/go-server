package nginx

import (
	// Namespace imports
	"io/fs"
	"path/filepath"
	"strings"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	// Available Configuration file path
	Available string
	// Enabled Configuration file path
	Enabled string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewConfig(available string) (*Config, error) {
	this := new(Config)

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Enumerates current set of files within a folder
func enumerate(root string, recursive bool) ([]string, error) {
	var result []string

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

		// Only enumerate regular files
		if info, err := d.Info(); err != nil {
			return nil
		} else if validFileMode(info.Mode()) {
			result = append(result, NewFile(path, info))
		}

		// Return success
		return nil
	}); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func validFileMode(mode fs.FileMode) bool {
	if mode.IsRegular() {
		return true
	}
	if mode.Type() == fs.ModeSymlink {
		return true
	}
	return false
}
