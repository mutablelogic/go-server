package parser

import (
	// Packages
	"fmt"
	"os"
	"path/filepath"
	"strings"

	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/mutablelogic/go-server/pkg/parser/jsonparser"
	meta "github.com/mutablelogic/go-server/pkg/parser/meta"
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
		} else if meta, err := meta.New(plugin, name); err != nil {
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
		r, err := os.Open(path)
		if err != nil {
			return httpresponse.ErrBadRequest.Withf("failed to open %q: %s", path, err)
		}
		defer r.Close()
		ast, err := jsonparser.Read(r)
		if err != nil {
			return httpresponse.ErrBadRequest.Withf("failed to parse %q: %s", path, err)
		}
		fmt.Println(ast)
	default:
		return httpresponse.ErrBadRequest.Withf("unsupported file extension: %q", ext)
	}

	// Return success
	return nil
}
