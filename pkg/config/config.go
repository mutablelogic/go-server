package config

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"

	// Modules
	multierror "github.com/hashicorp/go-multierror"
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
			} else if err := unmarshal(path, data, &r); err != nil {
				return fmt.Errorf("%s: %w", filepath.Base(path), err)
			} else if name := strings.TrimSpace(r.Name()); name == "" {
				return ErrBadParameter.Withf("%q: Resource has no name", d.Name())
			} else if !types.IsIdentifier(name) {
				return ErrBadParameter.Withf("%q Invalid resource with name: %q", d.Name(), name)
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
func LoadForResources(filesys fs.FS, resources []Resource, protos task.Plugins) (task.Plugins, error) {
	var result error
	var plugins = task.Plugins{}

	// Parse each file for resource definition, then create a new plugin with all the
	// configuration fields set
	for _, resource := range resources {
		name := resource.Name()
		path := resource.Path
		if proto, exists := protos[name]; !exists {
			result = multierror.Append(result, ErrNotFound.Withf("Plugin %q", name))
		} else if plugin := newPluginInstance(proto); plugin == nil {
			result = multierror.Append(result, ErrInternalAppError.Withf("LoadForResources: %q", name))
		} else if data, err := fs.ReadFile(filesys, path); err != nil {
			result = multierror.Append(result, err)
		} else if err := unmarshal(path, data, &plugin); err != nil {
			result = multierror.Append(result, fmt.Errorf("%s: %w", filepath.Base(path), err))
		} else if label := strings.TrimSpace(plugin.Label()); label == "" {
			result = multierror.Append(result, ErrBadParameter.Withf("%v: Resource has no label", filepath.Base(path)))
		} else if !types.IsIdentifier(label) {
			result = multierror.Append(result, ErrBadParameter.Withf("%v: Invalid resource with label: %q", filepath.Base(path), label))
		} else {
			key := name + "." + label
			if _, exists := plugins[key]; exists {
				result = multierror.Append(result, ErrBadParameter.Withf("%v: Duplicate resource with label: %q", filepath.Base(path), key))
			} else {
				plugins[key] = plugin
			}
		}
	}

	// Return result
	return plugins, result
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// unmarshal decodes data into a data structure. TODO: need to support
// formats other than JSON
func unmarshal(path string, data []byte, r any) error {
	if err := json.Unmarshal(data, r); err != nil {
		return err
	}
	fmt.Println("TODO: Resolve paths", path)
	return nil
}

// newPluginInstance returns a new plugin instance given a prototype
func newPluginInstance(proto Plugin) Plugin {
	if plugin, ok := reflect.New(reflect.TypeOf(proto)).Interface().(Plugin); ok {
		return plugin
	} else {
		return nil
	}
}
