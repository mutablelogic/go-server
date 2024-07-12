package provider

import (
	"errors"
	"fmt"
	"io"

	// Packages
	"github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/provider/ast"
	"github.com/mutablelogic/go-server/pkg/provider/dep"
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

// Append configurations from a JSON file
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

		// Handle var block
		label := config.Key()
		if label == labelVar {
			fmt.Println("TODO", label, config)
			continue
		}

		// Handle task block
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

type node struct {
	label types.Label
	path  types.Label
}

// Bind values
func (p *Parser) Bind() error {
	// Create a dependency graph and a root node
	dep := dep.NewGraph()
	root := &node{}

	// Evaluate each value, and create dependencies
	ctx := ast.NewContext(func(ctx *ast.Context, value any) (any, error) {
		if node, others, err := eval(ctx, value); err != nil {
			return nil, err
		} else {
			dep.AddNode(root, node)
			dep.AddNode(node, others...)
		}
		return nil, nil
	})

	// Traverse the nodes to create a dependency graph
	var result error
	for label, config := range p.configs {
		ctx.SetLabel(label)
		if _, err := config.Value(ctx); err != nil {
			result = errors.Join(result, err)
		}
	}
	if result != nil {
		return result
	}

	// Resolve the dependency graph
	resolved, err := dep.Resolve(root)
	if err != nil {
		return err
	}

	// Print out the resolved nodes
	for _, n := range resolved {
		fmt.Println(n.(*node).label, n.(*node).path, "=>")
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// evaluate value
func eval(ctx *ast.Context, value any) (dep.Node, []dep.Node, error) {
	label := ctx.Label()
	path := ctx.Path()
	switch value := value.(type) {
	case string:
		result := []dep.Node{}
		Expand(value, func(label string) string {
			fmt.Println("label", label)
			return ""
		})
		return &node{label, path}, result, nil
	default:
		return &node{label, path}, nil, nil
	}
}
