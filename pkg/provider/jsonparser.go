package provider

import (
	"encoding/json"
	"fmt"
	"io"

	// Packages
	"github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/provider/ast"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

type JSONParser struct {
	plugin map[string]server.Plugin
}

func NewJSONParser(plugins ...server.Plugin) (*JSONParser, error) {
	parser := new(JSONParser)
	parser.plugin = make(map[string]server.Plugin, len(plugins))
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

// Json parser state
type jstate int

const (
	_           jstate = iota
	sINIT              // Initial state
	sROOT              // Root state - top level {}
	sPLUGIN            // Plugin state - top level "plugin"
	sFIELD             // Plugin field name
	sFIELDVALUE        // Plugin field value
	sARRAYVALUE        // In an array
	sMAPKEY            // Map key
	sMAPVALUE          // Map value
)

func (s jstate) String() string {
	switch s {
	case sINIT:
		return "INIT"
	case sROOT:
		return "ROOT"
	case sPLUGIN:
		return "PLUGIN"
	case sFIELD:
		return "FIELD"
	case sFIELDVALUE:
		return "FIELDVALUE"
	case sARRAYVALUE:
		return "ARRAYVALUE"
	case sMAPKEY:
		return "MAPKEY"
	case sMAPVALUE:
		return "MAPVALUE"
	default:
		return fmt.Sprintf("Unknown(%v)", int(s))
	}
}

// Read a JSON file and return a root node
func (p *JSONParser) Read(r io.Reader) (ast.Node, error) {
	dec := json.NewDecoder(r)
	root := ast.NewRootNode()

	// Current node and state
	node, state := (ast.Node)(root), sINIT
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if t == nil {
			if node, state, err = parseNodeNil(node, state); err != nil {
				return nil, err
			}
			continue
		}

		// Debug
		fmt.Print("Token:", t, " Type:", fmt.Sprintf("%T", t), " State:", state, " Node:", node)

		switch t := t.(type) {
		case json.Delim:
			switch t {
			case '{':
				if node, state, err = parseNodeOpenBrace(node, state); err != nil {
					return nil, err
				}
			case '}':
				if node, state, err = parseNodeCloseBrace(node, state); err != nil {
					return nil, err
				}
			case '[':
				if node, state, err = parseNodeOpenSquare(node, state); err != nil {
					return nil, err
				}
			case ']':
				if node, state, err = parseNodeCloseSquare(node, state); err != nil {
					return nil, err
				}
			}
		case string:
			if node, state, err = parseNodeString(node, state, t); err != nil {
				return nil, err
			}
		case bool:
			if node, state, err = parseNodeBool(node, state, t); err != nil {
				return nil, err
			}
		case float64:
			if node, state, err = parseNodeFloat64(node, state, t); err != nil {
				return nil, err
			}
		default:
			return nil, ErrBadParameter.Withf("Unexpected token: %v (%T)", t, t)
		}

		fmt.Println(" =>", state)
	}

	// The end state should be -1
	if state != sINIT {
		return nil, ErrInternalAppError.With("Unexpected end state")
	}

	// Return success
	return root, nil
}

// Handle open brace
func parseNodeOpenBrace(node ast.Node, state jstate) (ast.Node, jstate, error) {
	switch state {
	case sINIT: // We start processing
		return node, sROOT, nil
	case sPLUGIN: // We are in a plugin
		return node, sFIELD, nil
	case sFIELDVALUE:
		// We are in an map
		return node.Append(ast.NewMapNode(node)), sMAPKEY, nil
	case sMAPVALUE: // We are in a map
		return node.Append(ast.NewMapNode(node)), sMAPKEY, nil
	case sARRAYVALUE: // We are in a array
		return node.Append(ast.NewMapNode(node)), sMAPKEY, nil
	}
	return node, state, ErrBadParameter.With("Unexpected '{'")
}

// Handle close brace
func parseNodeCloseBrace(node ast.Node, state jstate) (ast.Node, jstate, error) {
	switch state {
	case sROOT:
		// End of file
		return node, sINIT, nil
	case sPLUGIN:
		// End of plugin
		if node.Parent().Type() == ast.Root {
			return node, sROOT, nil
		} else {
			return node, state, ErrBadParameter.With("Unexpected parent ", node.Parent())
		}
	case sFIELD:
		// End of field
		if node.Parent().Type() == ast.Plugin {
			return node.Parent(), sPLUGIN, nil
		}
		return node, state, ErrBadParameter.With("Unexpected parent ", node.Parent())
	case sMAPVALUE:
		// End of map - field
		if node.Parent().Type() == ast.Field {
			return node.Parent(), sFIELD, nil
		}
		// End of map - array
		if node.Parent().Type() == ast.Array {
			return node.Parent(), sARRAYVALUE, nil
		}
		// End of map - map
		if node.Parent().Type() == ast.Map {
			return node.Parent(), sMAPKEY, nil
		}
		return node, state, ErrBadParameter.With("Unexpected parent ", node.Parent())
	}
	return node, state, ErrBadParameter.With("Unexpected '}'")
}

// Handle open square bracket
func parseNodeOpenSquare(node ast.Node, state jstate) (ast.Node, jstate, error) {
	switch state {
	case sFIELDVALUE:
		// We are in an array
		return node.Append(ast.NewArrayNode(node)), sARRAYVALUE, nil
	case sARRAYVALUE:
		// We are in an array
		return node.Append(ast.NewArrayNode(node)), sARRAYVALUE, nil
	}
	return node, state, ErrBadParameter.With("Unexpected '['")
}

// Handle close square bracket
func parseNodeCloseSquare(node ast.Node, state jstate) (ast.Node, jstate, error) {
	switch state {
	case sARRAYVALUE:
		// End of array - field
		if node.Parent().Type() == ast.Field {
			return node.Parent(), sFIELD, nil
		}
		// End of array - array
		if node.Parent().Type() == ast.Array {
			return node.Parent(), sARRAYVALUE, nil
		}
		// End of array - map
		if node.Parent().Type() == ast.Map {
			return node.Parent(), sMAPKEY, nil
		}
		return node, state, ErrBadParameter.With("Unexpected parent ", node.Parent())
	}
	return node, state, ErrBadParameter.With("Unexpected ']'")
}

// Handle string
func parseNodeString(node ast.Node, state jstate, value string) (ast.Node, jstate, error) {
	switch state {
	case sROOT:
		// Plugin name
		return node.Append(ast.NewPluginNode(node, value)), sPLUGIN, nil
	case sFIELD:
		// Plugin field name
		return node.Append(ast.NewFieldNode(node, value)), sFIELDVALUE, nil
	case sFIELDVALUE:
		// Plugin field value
		node.Append(ast.NewValueNode(node, nil))
		return node.Parent(), sFIELD, nil
	case sMAPKEY:
		// Map field key
		node.Append(ast.NewValueNode(node, value))
		return node, sMAPVALUE, nil
	case sMAPVALUE:
		// Map field value
		node.Append(ast.NewValueNode(node, value))
		return node.Parent(), sMAPKEY, nil
	case sARRAYVALUE:
		// Array value
		node.Append(ast.NewValueNode(node, value))
		return node, sARRAYVALUE, nil
	}
	return node, state, ErrBadParameter.Withf("Unexpected %q", value)
}

// Handle bool
func parseNodeBool(node ast.Node, state jstate, value bool) (ast.Node, jstate, error) {
	switch state {
	case sFIELDVALUE:
		// Plugin field value
		node.Append(ast.NewValueNode(node, value))
		return node.Parent(), sFIELD, nil
	case sMAPVALUE:
		// Map field value
		node.Append(ast.NewValueNode(node, value))
		return node.Parent(), sMAPKEY, nil
	case sARRAYVALUE:
		// Array value
		node.Append(ast.NewValueNode(node, value))
		return node, sARRAYVALUE, nil
	}
	return node, state, ErrBadParameter.Withf("Unexpected '%v'", value)
}

// Handle float64
func parseNodeFloat64(node ast.Node, state jstate, value float64) (ast.Node, jstate, error) {
	switch state {
	case sFIELDVALUE:
		// Plugin field value
		node.Append(ast.NewValueNode(node, value))
		return node.Parent(), sFIELD, nil
	case sMAPVALUE:
		// Map field value
		node.Append(ast.NewValueNode(node, value))
		return node.Parent(), sMAPKEY, nil
	case sARRAYVALUE:
		// Array value
		node.Append(ast.NewValueNode(node, value))
		return node, sARRAYVALUE, nil
	}
	return node, state, ErrBadParameter.Withf("Unexpected '%v'", value)
}

// Handle null
func parseNodeNil(node ast.Node, state jstate) (ast.Node, jstate, error) {
	switch state {
	case sFIELDVALUE:
		// Plugin field value
		node.Append(ast.NewValueNode(node, nil))
		return node.Parent(), sFIELD, nil
	case sMAPVALUE:
		// Map field value
		node.Append(ast.NewValueNode(node, nil))
		return node.Parent(), sMAPKEY, nil
	case sARRAYVALUE:
		// Array value
		node.Append(ast.NewValueNode(node, nil))
		return node, sARRAYVALUE, nil
	}
	return node, state, ErrBadParameter.With("Unexpected nil token")
}
