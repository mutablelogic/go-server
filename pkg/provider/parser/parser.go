package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	meta "github.com/mutablelogic/go-server/pkg/provider/meta"
	ast "github.com/mutablelogic/go-server/pkg/provider/parser/ast"
	jsonparser "github.com/mutablelogic/go-server/pkg/provider/parser/jsonparser"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type parser struct {
	meta      map[string]*meta.Meta
	variables map[string]*Variable
	resources map[string]*Resource
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new parser with the given plugins
func New(plugins ...server.Plugin) (*parser, error) {
	p := new(parser)
	p.meta = make(map[string]*meta.Meta, len(plugins))
	p.resources = make(map[string]*Resource, len(plugins)*2)
	p.variables = make(map[string]*Variable, len(plugins))

	// Read all plugins, and create metadata for them
	for _, plugin := range plugins {
		name := plugin.Name()
		if _, exists := p.meta[name]; exists {
			return nil, httpresponse.ErrConflict.Withf("duplicate plugin: %q", name)
		} else if meta, err := meta.New(plugin); err != nil {
			return nil, httpresponse.ErrInternalError.Withf("%s for %q", err, name)
		} else {
			p.meta[name] = meta
		}
	}

	// Return success
	return p, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Read a configuration file and parse it, creating variables and resources
func (p *parser) Parse(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		// Read the file
		r, err := os.Open(path)
		if err != nil {
			return httpresponse.ErrBadRequest.Withf("failed to open %q: %s", path, err)
		}
		defer r.Close()

		// Create the syntax tree
		ast, err := jsonparser.Read(r)
		if err != nil {
			return httpresponse.ErrBadRequest.Withf("failed to parse %q: %s", path, err)
		} else if err := p.jsonParse(ast); err != nil {
			return httpresponse.ErrBadRequest.Withf("failed to parse %q: %s", path, err)
		}
	default:
		return httpresponse.ErrBadRequest.Withf("unsupported file extension: %q", ext)
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (p *parser) jsonParse(root ast.Node) error {
	if root.Type() != ast.Dict {
		return httpresponse.ErrBadRequest.Withf("expected object, got %s", root.Type())
	}

	// tree should contain "variable" = <dict> or "<resource>" = <dict>
	for _, ident := range root.Children() {
		if ident.Type() != ast.Ident {
			return httpresponse.ErrBadRequest.Withf("expected identifier, got %s", ident.Type())
		}

		name, ok := ident.Value().(string)
		if !ok || name == "" {
			return httpresponse.ErrBadRequest.Withf("expected identifier, got %q", ident.Value())
		}

		// Ignore comments
		if name == kwComment {
			continue
		}

		// Create variable
		if name == kwVariable && len(ident.Children()) > 0 {
			if err := p.jsonVariablesParse(ident.Children()[0]); err != nil {
				return httpresponse.ErrBadRequest.Withf("failed to parse %q: %s", name, err)
			}
			continue
		}

		// Resource with metadata
		if meta, exists := p.meta[name]; exists && len(ident.Children()) > 0 {
			if err := p.jsonPluginsParse(meta, ident.Children()[0]); err != nil {
				return httpresponse.ErrBadRequest.Withf("failed to parse %q: %s", name, err)
			}
			continue
		}

		// Nothing found
		return httpresponse.ErrNotFound.Withf("resource not found: %q", name)
	}

	// Return success
	return nil
}

func (p *parser) jsonVariablesParse(root ast.Node) error {
	if root.Type() != ast.Dict {
		return httpresponse.ErrBadRequest.Withf("expected object, got %s", root.Type())
	}

	// tree should contain "<variable>" = <dict>
	for _, ident := range root.Children() {
		if ident.Type() != ast.Ident {
			return httpresponse.ErrBadRequest.Withf("expected identifier, got %s", ident.Type())
		}

		name, ok := ident.Value().(string)
		if !ok || name == "" {
			return httpresponse.ErrBadRequest.Withf("expected identifier, got %q", ident.Value())
		} else if _, exists := p.variables[name]; exists {
			return httpresponse.ErrBadRequest.Withf("duplicate variable: %q", name)
		} else if len(ident.Children()) > 0 {
			if variable, err := jsonVariableParse(name, ident.Children()[0]); err != nil {
				return httpresponse.ErrBadRequest.Withf("failed to parse %q: %s", name, err)
			} else {
				p.variables[name] = variable
			}
		}
	}

	// Return success
	return nil
}

func (p *parser) jsonPluginsParse(meta *meta.Meta, root ast.Node) error {
	if root.Type() != ast.Dict {
		return httpresponse.ErrBadRequest.Withf("expected object, got %s", root.Type())
	}

	// tree should contain "<label>" = <dict>
	for _, ident := range root.Children() {
		if ident.Type() != ast.Ident {
			return httpresponse.ErrBadRequest.Withf("expected label, got %s", ident.Type())
		}

		label, ok := ident.Value().(string)
		if !ok || label == "" {
			return httpresponse.ErrBadRequest.Withf("expected label, got %q", ident.Value())
		} else {
			label = meta.Label()
		}

		if _, exists := p.resources[label]; exists {
			return httpresponse.ErrBadRequest.Withf("duplicate resource: %q", label)
		} else if len(ident.Children()) > 0 {
			if resource, err := jsonResourceParse(meta, label, ident.Children()[0]); err != nil {
				return httpresponse.ErrBadRequest.Withf("failed to parse %q: %s", label, err)
			} else {
				p.resources[label] = resource
			}
		}
	}
	return nil
}

func jsonVariableParse(name string, root ast.Node) (*Variable, error) {
	if root.Type() != ast.Dict {
		return nil, httpresponse.ErrBadRequest.Withf("expected object, got %s", root.Type())
	}

	// tree should contain "description" and "default" only
	v := new(Variable)
	v.Name = name
	for _, ident := range root.Children() {
		if ident.Type() != ast.Ident {
			return nil, httpresponse.ErrBadRequest.Withf("expected identifier, got %s", ident.Type())
		} else if len(ident.Children()) == 0 {
			continue
		}

		name, ok := ident.Value().(string)
		if !ok || name == "" {
			return nil, httpresponse.ErrBadRequest.Withf("expected identifier, got %q", ident.Value())
		}
		value := ident.Children()[0]

		switch name {
		case kwDescription:
			if value.Type() != ast.String {
				return nil, httpresponse.ErrBadRequest.Withf("expected string, got %s", value.Type())
			} else {
				v.Description = value.Value().(string)
			}
		case kwDefault:
			v.Default = value.Value()
		default:
			return nil, httpresponse.ErrBadRequest.Withf("expected description or default, got %q", ident.Value())
		}
	}
	return v, nil
}

func jsonResourceParse(meta *meta.Meta, label string, root ast.Node) (*Resource, error) {
	if root.Type() != ast.Dict {
		return nil, httpresponse.ErrBadRequest.Withf("expected object, got %s", root.Type())
	}

	// tree should contain key/value pairs and objects which may be nested
	r := new(Resource)
	r.Meta = meta
	r.Label = label

	for _, ident := range root.Children() {
		if ident.Type() != ast.Ident {
			return nil, httpresponse.ErrBadRequest.Withf("expected identifier, got %s", ident.Type())
		} else if len(ident.Children()) == 0 {
			continue
		}

		name, ok := ident.Value().(string)
		if !ok || name == "" {
			return nil, httpresponse.ErrBadRequest.Withf("expected identifier, got %q", ident.Value())
		}
		value := ident.Children()[0]

		fmt.Println("name:", name, "value:", value)

	}
	return r, nil
}
