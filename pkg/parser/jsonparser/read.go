package jsonparser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	// Packages
	ast "github.com/mutablelogic/go-server/pkg/parser/ast"
)

func Read(r io.Reader) (ast.Node, error) {
	var root, cur ast.Node

	parser := json.NewDecoder(r)
	parser.UseNumber() // Use Number to preserve number types

	// Iterate over the tokens until EOF
	for {
		token, err := parser.Token()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
		switch token := token.(type) {
		case string:
			cur, err = parseString(cur, token)
			if err != nil {
				return nil, err
			}
		case nil:
			cur, err = parseNull(cur)
			if err != nil {
				return nil, err
			}
		case bool:
			cur, err = parseBool(cur, token)
			if err != nil {
				return nil, err
			}
		case json.Number:
			cur, err = parseNumber(cur, string(token))
			if err != nil {
				return nil, err
			}
		case json.Delim:
			cur, err = parseDelim(cur, rune(token))
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unexpected token type: %T", token)
		}
		// Set root to the first node
		if root == nil {
			root = cur
		}
	}

	// Return the root node
	return root, nil
}

func parseString(parent ast.Node, token string) (ast.Node, error) {
	if parent == nil {
		return ast.NewString(nil, token), nil
	}
	switch parent.Type() {
	case ast.Dict:
		return ast.NewIdent(parent, token), nil
	case ast.Array:
		return ast.NewString(parent, token), nil
	case ast.Ident:
		ast.NewString(parent, string(token))
		return parent.Parent(), nil
	default:
		return nil, fmt.Errorf("unexpected string: %q", token)
	}
}

func parseNull(parent ast.Node) (ast.Node, error) {
	if parent == nil {
		return ast.NewNull(nil), nil
	}
	switch parent.Type() {
	case ast.Array:
		ast.NewNull(parent)
		return parent, nil
	case ast.Ident:
		ast.NewNull(parent)
		return parent.Parent(), nil
	default:
		return nil, fmt.Errorf("unexpected null")
	}
}

func parseBool(parent ast.Node, token bool) (ast.Node, error) {
	if parent == nil {
		return ast.NewBool(nil, token), nil
	}
	switch parent.Type() {
	case ast.Array:
		ast.NewBool(parent, token)
		return parent, nil
	case ast.Ident:
		ast.NewBool(parent, token)
		return parent.Parent(), nil
	default:
		return nil, fmt.Errorf("unexpected bool: %v", token)
	}
}

func parseNumber(parent ast.Node, token string) (ast.Node, error) {
	if parent == nil {
		return ast.NewNumber(nil, token), nil
	}
	switch parent.Type() {
	case ast.Array:
		ast.NewNumber(parent, token)
		return parent, nil
	case ast.Ident:
		ast.NewNumber(parent, token)
		return parent.Parent(), nil
	default:
		return nil, fmt.Errorf("unexpected number: %v", token)
	}
}

func parseDelim(parent ast.Node, token rune) (ast.Node, error) {
	switch token {
	case '{':
		return ast.NewDict(parent), nil
	case '}':
		switch parent.Type() {
		case ast.Dict:
			return parent.Parent(), nil
		case ast.Ident:
			return parent.Parent(), nil
		default:
			return nil, fmt.Errorf("unexpected token: %q for %s", token, parent.Type())
		}
	case '[':
		return ast.NewArray(parent), nil
	case ']':
		switch parent.Type() {
		case ast.Array:
			return parent.Parent(), nil
		}
	}
	return nil, fmt.Errorf("unexpected token: %q", token)
}
