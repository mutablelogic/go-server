package config

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"

	// Modules
	json "github.com/mutablelogic/go-server/pkg/config/json"
	task "github.com/mutablelogic/go-server/pkg/task"
	types "github.com/mutablelogic/go-server/pkg/types"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
)

/////////////////////////////////////////////////////////////////////
// TYPES

// Resource is a plugin resource definition with name and label
type Resource struct {
	task.Plugin
	Path string `json:"-"`
}

/////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	fileExtJson = ".json"
)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// LoadForPattern returns resources for a pattern of files on a filesystem, in
// no particular order. Currently only JSON files are supported, with a .json
// extension or else an error is returned
func LoadForPattern(filesys fs.FS, pattern string) ([]Resource, error) {
	var result error
	resources := []Resource{}

	// Glob
	paths, err := fs.Glob(filesys, pattern)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, ErrNotFound.Withf(pattern)
	}

	// Parse each file for resource definition
	for _, path := range paths {
		if err := fs.WalkDir(filesys, path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if strings.HasPrefix(d.Name(), ".") {
					return fs.SkipDir
				} else {
					return nil
				}
			}
			if !d.Type().IsRegular() {
				return ErrBadParameter.Withf("Not a regular file: %q", d.Name())
			}
			// Read the resource
			r := Resource{Path: path}
			if data, err := fs.ReadFile(filesys, path); err != nil {
				return err
			} else if err := unmarshal(data, &r); err != nil {
				return err
			} else if name := strings.TrimSpace(r.Name()); name == "" {
				return ErrBadParameter.Withf("%q: Resource has no name", d.Name())
			} else if !types.IsIdentifier(name) {
				return ErrBadParameter.Withf("%q: Invalid resource with name: %q", path, name)
			} else {
				resources = append(resources, r)
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	// Return result
	return resources, result
}

// LoadForResources returns plugins for a list of resources
func LoadForResources(filesys fs.FS, resources []Resource, plugins task.Plugins) ([]Resource, error) {
	// Parse each file for resource definition
	for _, resource := range resources {
		// Create a new plugin
		name := resource.Name()
		if plugin, exists := plugins[name]; !exists {
			return nil, ErrNotFound.Withf("Plugin %q", name)
		} else if plugin_ := newPluginInstance(plugin); plugin_ == nil {
			return nil, ErrBadParameter.Withf("Plugin %q is not a resource", name)
		} else {
			fmt.Println("LOAD", resource.Path, "PLUGIN", plugin_)
		}
	}

	// Return result
	return resources, nil
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

*/

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func unmarshal(data []byte, r *Resource) error {
	ext := strings.ToLower(filepath.Ext(r.Path))
	switch ext {
	case fileExtJson:
		return json.Unmarshal(data, r)
	default:
		return ErrBadParameter.Withf("%q: Unsupported file extension", filepath.Base(r.Path))
	}
}

// newPluginInstance returns a new plugin instance given a prototype
func newPluginInstance(plugin Plugin) Plugin {
	return reflect.New(reflect.TypeOf(plugin)).Interface().(task.Plugin)
}
