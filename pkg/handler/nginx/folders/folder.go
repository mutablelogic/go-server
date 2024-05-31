package folders

import (
	"encoding/json"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Folder struct {
	// Root path
	Root string

	// Map hash of file contents to files
	Files map[string]*File
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func NewFolder(root string, ext string, recursive bool) (*Folder, error) {
	this := new(Folder)

	// Enumerate files
	files, err := enumerate(root, ext, recursive)
	if err != nil {
		return nil, err
	} else {
		this.Root = root
	}

	// Create map of files, silently ignoring any files
	// with the same hash or that could not be read
	this.Files = make(map[string]*File, len(files))
	for _, file := range files {
		if _, exists := this.Files[file.Hash]; exists {
			continue
		}
		this.Files[file.Hash] = file
	}

	// Return success
	return this, nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (f Folder) String() string {
	data, _ := json.MarshalIndent(f, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Enumerates current set of files within a folder
func enumerate(root, ext string, recursive bool) ([]*File, error) {
	var result []*File

	ext = strings.ToLower(ext)
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

		// Check file extension
		if ext != "" && strings.ToLower(filepath.Ext(path)) != ext {
			return nil
		}
		// Only enumerate regular files, ignore any files with errors
		info, err := d.Info()
		if err != nil {
			return nil
		}
		// Check filenames
		if !types.IsFilename(info.Name()) {
			return nil
		}
		// Check file type and size
		if !info.Mode().IsRegular() || info.Size() == 0 {
			return nil
		}

		// Read the file
		if relpath, err := filepath.Rel(root, path); err == nil {
			if file, err := NewFile(root, relpath); err == nil {
				result = append(result, file)
			}
		}

		// Return success
		return nil
	}); err != nil {
		return nil, err
	}

	// Return success
	return result, nil
}
