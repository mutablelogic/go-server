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
		if name == labelVar {
			return nil, ErrBadParameter.Withf("plugin cannot be named %q", name)
		} else if _, exists := parser.plugin[name]; exists {
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
	tree, err := json.Parse(r, nil)
	if err != nil {
		return err
	}

	// Get the plugin configurations with no evaluation
	configs := tree.Children()
	for _, config := range configs {
		if config.Type() != ast.Value {
			result = errors.Join(result, ErrInternalAppError.Withf("unexpected type %v", config.Type()))
			continue
		}

		label := config.Key()
		if label == labelVar {
			fmt.Println("TODO", label)
			continue
		}

		// Parse a task block
		if label, err := types.ParseLabel(label); err != nil {
			result = errors.Join(result, err)
		} else if _, exists := p.configs[label]; exists {
			result = errors.Join(result, ErrDuplicateEntry.Withf("Duplicate label %q", label))
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

	// Set the evaluation function
	ctx := ast.NewContext(func(ctx *ast.Context, value any) (any, error) {
		fmt.Printf("EVAL %s.%s => %q (%T)\n", ctx.Label(), ctx.Path(), value, value)
		return nil, nil
	})

	// Process each configuration in order to discover the dependencies
	for label, config := range p.configs {
		ctx.SetLabel(label)
		_, err := config.Value(ctx)
		if err != nil {
			result = errors.Join(result, err)
		}
	}

	// Return any errors
	return result
}
