package parser

import (
	"os"
	"path/filepath"
	"strings"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	ast "github.com/mutablelogic/go-server/pkg/parser/ast"
	jsonparser "github.com/mutablelogic/go-server/pkg/parser/jsonparser"
	meta "github.com/mutablelogic/go-server/pkg/provider/meta"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type parser struct {
	meta map[string]*meta.Meta
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(plugins ...server.Plugin) (*parser, error) {
	p := new(parser)
	p.meta = make(map[string]*meta.Meta, len(plugins))

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

// Read a file and parse it
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

	// tree should contain <dict plugin:<dict values> plugin:<dict values> ...>
	for _, plugins := range root.Children() {
		if plugins.Type() != ast.Ident {
			return httpresponse.ErrBadRequest.Withf("expected identifier, got %s", plugins.Type())
		}
		plugin := plugins.Value().(string)
		if plugin_, exists := p.meta[plugin]; !exists {
			return httpresponse.ErrNotFound.Withf("plugin not found: %q", plugin)
		} else if children := plugins.Children(); len(children) != 1 {
			return httpresponse.ErrBadRequest.Withf("expected one child, got %d", len(children))
		} else if dict := children[0]; dict.Type() != ast.Dict {
			return httpresponse.ErrBadRequest.Withf("expected object, got %s", dict.Type())
		} else if err := plugin_.Validate(dict.Value()); err != nil {
			return err
		}
	}

	// Return success
	return nil
}
