package json

import (
	"encoding/json"
	"fmt"
	"io"

	// Packages
	"github.com/mutablelogic/go-server/pkg/provider/ast"
	"github.com/mutablelogic/go-server/pkg/types"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Json parser state
type jstate int

const (
	_         jstate = iota
	sINIT            // Initial state
	sARRAY           // In an array
	sMAPKEY          // Map key
	sMAPVALUE        // Map value
)

// Arguments are token, state and node
type DebugFunc func(any, string, ast.Node)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s jstate) String() string {
	switch s {
	case sINIT:
		return "INIT"
	case sARRAY:
		return "ARRAY"
	case sMAPKEY:
		return "MAPKEY"
	case sMAPVALUE:
		return "MAPVALUE"
	default:
		return fmt.Sprintf("Unknown(%v)", int(s))
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Read a JSON file and return a root node
func Parse(r io.Reader, fn DebugFunc) (ast.Node, error) {
	dec := json.NewDecoder(r)
	dec.UseNumber()

	// Root node, current node and state
	root := ast.NewMapNode(nil)
	node, state := (ast.Node)(root), sINIT
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if fn != nil {
			fn(t, state.String(), node)
		}

		if t == nil {
			if node, state, err = parseNodeNil(node, state); err != nil {
				return nil, err
			}
			continue
		}

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
		case json.Number:
			if node, state, err = parseNodeNumber(node, state, t); err != nil {
				return nil, err
			}
		default:
			return nil, ErrBadParameter.Withf("Unexpected token: %v (%T)", t, t)
		}
	}

	// The end state should be a node with nil value
	if node != nil {
		return root, ErrInternalAppError.With("Unexpected end state")
	}

	// Return success
	return root, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Handle open brace
func parseNodeOpenBrace(node ast.Node, state jstate) (ast.Node, jstate, error) {
	switch state {
	case sINIT: // We are at the top level
		return node, sMAPKEY, nil
	case sMAPVALUE: // We are in a map
		return node.Append(ast.NewMapNode(node)), sMAPKEY, nil
	case sARRAY: // We are in a array
		return node.Append(ast.NewMapNode(node)), sMAPKEY, nil
	}
	return node, state, ErrBadParameter.With("Unexpected '{'")
}

// Handle close brace
func parseNodeCloseBrace(node ast.Node, state jstate) (ast.Node, jstate, error) {
	switch state {
	case sMAPKEY:
		// End of map - top level
		if node.Parent() == nil {
			return nil, sINIT, nil
		}
		// End of map - array
		if node.Parent().Type() == ast.Array {
			return node.Parent(), sARRAY, nil
		}
		// End of map - map
		if node.Type() == ast.Map {
			return node.Parent().Parent(), sMAPKEY, nil
		}
		return node, state, ErrBadParameter.With("Unexpected parent ", node.Parent())
	}
	return node, state, ErrBadParameter.With("Unexpected '}'")
}

// Handle open square bracket
func parseNodeOpenSquare(node ast.Node, state jstate) (ast.Node, jstate, error) {
	switch state {
	case sARRAY:
		// We are in an array
		return node.Append(ast.NewArrayNode(node)), sARRAY, nil
	case sMAPVALUE:
		// We are in a map
		return node.Append(ast.NewArrayNode(node)), sARRAY, nil
	}
	return node, state, ErrBadParameter.With("Unexpected '['")
}

// Handle close square bracket
func parseNodeCloseSquare(node ast.Node, state jstate) (ast.Node, jstate, error) {
	switch state {
	case sARRAY:
		// End of array - array
		if node.Parent().Type() == ast.Array {
			return node.Parent(), sARRAY, nil
		}
		// End of array - map
		if node.Parent().Parent().Type() == ast.Map {
			return node.Parent().Parent(), sMAPKEY, nil
		}
		return node, state, ErrBadParameter.With("Unexpected parent ", node.Parent())
	}
	return node, state, ErrBadParameter.With("Unexpected ']'")
}

// Handle string
func parseNodeString(node ast.Node, state jstate, value string) (ast.Node, jstate, error) {
	switch state {
	case sMAPKEY:
		// Check key type - at top level, key must be a label or else it must be an identifier
		if node.Parent() == nil {
			_, err := types.ParseLabel(value)
			if err != nil {
				return node, state, err
			}
		} else if !types.IsIdentifier(value) {
			return node, state, ErrBadParameter.Withf("Invalid map key %q %v", value, node.Parent())
		}

		// Map field key - add a value
		child, err := ast.NewMapValueNode(node, value)
		if err != nil {
			return node, state, err
		}
		return node.Append(child), sMAPVALUE, nil
	case sMAPVALUE:
		if node.Type() == ast.Value {
			node.Append(ast.NewValueNode(node, value))
			return node.Parent(), sMAPKEY, nil
		}
		return node, state, ErrBadParameter.With("Unexpected parent ", node.Parent())
	case sARRAY:
		// Array value
		node.Append(ast.NewValueNode(node, value))
		return node, sARRAY, nil
	}
	return node, state, ErrBadParameter.Withf("Unexpected %q", value)
}

// Handle bool
func parseNodeBool(node ast.Node, state jstate, value bool) (ast.Node, jstate, error) {
	switch state {
	case sMAPVALUE:
		if node.Type() == ast.Value {
			node.Append(ast.NewValueNode(node, value))
			return node.Parent(), sMAPKEY, nil
		}
		return node, state, ErrBadParameter.With("Unexpected parent ", node.Parent())
	case sARRAY:
		// Array value
		node.Append(ast.NewValueNode(node, value))
		return node, sARRAY, nil
	}
	return node, state, ErrBadParameter.Withf("Unexpected '%v'", value)
}

// Handle json.Number
func parseNodeNumber(node ast.Node, state jstate, value json.Number) (ast.Node, jstate, error) {
	switch state {
	case sMAPVALUE:
		if node.Type() == ast.Value {
			node.Append(ast.NewValueNode(node, value))
			return node.Parent(), sMAPKEY, nil
		}
		return node, state, ErrBadParameter.With("Unexpected parent ", node.Parent())
	case sARRAY:
		// Array value
		node.Append(ast.NewValueNode(node, value))
		return node, sARRAY, nil
	}
	return node, state, ErrBadParameter.Withf("Unexpected '%v'", value)
}

// Handle null
func parseNodeNil(node ast.Node, state jstate) (ast.Node, jstate, error) {
	switch state {
	case sMAPVALUE:
		if node.Type() == ast.Value {
			node.Append(ast.NewValueNode(node, nil))
			return node.Parent(), sMAPKEY, nil
		}
		return node, state, ErrBadParameter.With("Unexpected parent ", node.Parent())
	case sARRAY:
		// Array value
		node.Append(ast.NewValueNode(node, nil))
		return node, sARRAY, nil
	}
	return node, state, ErrBadParameter.With("Unexpected nil token")
}
