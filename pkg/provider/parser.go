package provider

import (
	"errors"
	"fmt"
	"io"

	// Packages
	"github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/provider/ast"
	"github.com/mutablelogic/go-server/pkg/provider/json"
	"github.com/mutablelogic/go-server/pkg/types"

	// Namesoace imports
	. "github.com/djthorpe/go-errors"
)

type Parser struct {
	plugin  map[string]server.Plugin
	configs map[types.Label]ast.Node
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewParser(plugins ...server.Plugin) (*Parser, error) {
	parser := new(Parser)
	parser.plugin = make(map[string]server.Plugin, len(plugins))
	parser.configs = make(map[types.Label]ast.Node, len(plugins)*2)

	for _, plugin := range plugins {
		name := plugin.Name()
		if _, exists := parser.plugin[name]; exists {
			return nil, ErrDuplicateEntry.Withf("plugin %q already exists", plugin.Name())
		} else {
			parser.plugin[name] = plugin
		}
	}

	// Return success
	return parser, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Append an ast.Tree node
func (p *Parser) ParseJSON(r io.Reader) error {
	var result error

	// Parse JSON into a tree
	tree, err := json.Parse(r)
	if err != nil {
		return err
	}

	// Get the plugin configurations with no evaluation
	configs := tree.Children()
	for _, config := range configs {
		if config.Type() != ast.Plugin {
			result = errors.Join(result, ErrInternalAppError.Withf("%v", config.Type()))
			continue
		}
		if label, err := types.ParseLabel(config.Key()); err != nil {
			result = errors.Join(result, err)
		} else if _, exists := p.configs[label]; exists {
			result = errors.Join(result, ErrDuplicateEntry.Withf("%q", label))
		} else if _, exists := p.plugin[label.Prefix()]; !exists {
			result = errors.Join(result, ErrNotFound.Withf("Plugin %q", label.Prefix()))
		} else {
			p.configs[label] = config
		}
	}

	// Return any errors
	return result
}

// Bind values
func (p *Parser) Bind() error {
	var result error

	ctx := ast.NewContext(func(ctx *ast.Context, value any) (any, error) {
		fmt.Println("EVAL", value)
		return nil, nil
	})
	for _, config := range p.configs {
		_, err := config.Value(ctx)
		if err != nil {
			result = errors.Join(result, err)
		}
	}

	// Return any errors
	return result
}
