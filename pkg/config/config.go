package config

import (
	"io/fs"
	"path/filepath"
	"strings"

	// Modules
	multierror "github.com/hashicorp/go-multierror"
	json "github.com/mutablelogic/go-server/pkg/config/json"
	task "github.com/mutablelogic/go-server/pkg/task"
	types "github.com/mutablelogic/go-server/pkg/types"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	//. "github.com/mutablelogic/go-server"
)

/////////////////////////////////////////////////////////////////////
// TYPES

// Resource is a plugin resource definition with name and label
type Resource struct {
	task.Plugin
	Path string `json:"-"`
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// LoadForPattern returns resources for a pattern of files on a filesystem, in
// no particular order. Currently only JSON files are supported, with a .json
// extension or else an error is returned
func LoadForPattern(filesys fs.FS, pattern string) ([]Resource, error) {
	var result error
	resources := []Resource{}

	// Find files
	files, err := fs.Glob(filesys, pattern)
	if err != nil {
		return nil, err
	}

	// Parse each file for resource definition
	for _, path := range files {
		r := Resource{Path: path}
		if data, err := fs.ReadFile(filesys, path); err != nil {
			result = multierror.Append(result, err)
		} else if err := unmarshal(data, &r); err != nil {
			result = multierror.Append(result, err)
		} else if name := strings.TrimSpace(r.Name()); name == "" {
			result = multierror.Append(result, ErrBadParameter.Withf("%q: Resource has no name", path))
		} else if !types.IsIdentifier(name) {
			result = multierror.Append(result, ErrBadParameter.Withf("%q: Invalid resource with name: %q", path, name))
		} else {
			resources = append(resources, r)
		}
	}

	// Return result
	return resources, result
}

/*
// ParseJSONResource will return a Plugin given a resource and a prototype
func ParseJSONResource(filesys fs.FS, resource Resource, protos task.Plugins) (Plugin, error) {
	plugin := newPluginInstance(protos, resource.Name)
	if data, err := fs.ReadFile(filesys, resource.Path); err != nil {
		return nil, err
	} else if err := json.Unmarshal(data, plugin); err != nil {
		return nil, err
	}

	// Return success
	return plugin, nil
}

func newPluginInstance(plugin Plugin) Task {
	rv := reflect.New(reflect.TypeOf(plugin))
	return rv.Interface().(Plugin)
}

*/

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func unmarshal(data []byte, r *Resource) error {
	ext := filepath.Ext(r.Path)
	switch ext {
	case ".json":
		return json.Unmarshal(data, r)
	default:
		return ErrBadParameter.Withf("%q: Unsupported file extension: %q", r.Path, ext)
	}
}
