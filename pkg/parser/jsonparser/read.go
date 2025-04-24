package jsonparser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/mutablelogic/go-server/pkg/parser/ast"
)

func Read(r io.Reader) (ast.Node, error) {
	parser := json.NewDecoder(r)
	parser.UseNumber() // Use Number to preserve number types

	// Root element is a block, which will contain other blocks
	root := ast.NewBlock(nil)
	cur := root

	for {
		token, err := parser.Token()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
		switch token := token.(type) {
		case string:
			switch cur.Type() {
			case ast.Block:
				cur = ast.NewIdent(cur, token)
			case ast.Ident:
				ast.NewString(cur, string(token))
				cur = cur.Parent()
			default:
				return nil, fmt.Errorf("unexpected string token: %q", token)
			}
		case nil:
			switch cur.Type() {
			case ast.Ident:
				ast.NewNull(cur)
				cur = cur.Parent()
			default:
				return nil, fmt.Errorf("unexpected null token: %q", token)
			}
		case bool:
			switch cur.Type() {
			case ast.Ident:
				ast.NewBool(cur, token)
				cur = cur.Parent()
			default:
				return nil, fmt.Errorf("unexpected bool token: %v", token)
			}
		case json.Number:
			switch cur.Type() {
			case ast.Ident:
				ast.NewNumber(cur, string(token))
				cur = cur.Parent()
			case ast.Array:
				ast.NewNumber(cur, string(token))
			default:
				return nil, fmt.Errorf("unexpected number token: %q", token)
			}
		case json.Delim:
			switch token {
			case '{':
				cur = ast.NewBlock(cur)
			case '}':
				switch cur.Type() {
				case ast.Block:
					cur = cur.Parent()
				default:
					return nil, fmt.Errorf("unexpected block close token: %q", token)
				}
			case '[':
				switch cur.Type() {
				case ast.Ident:
					cur = ast.NewArray(cur)
				default:
					return nil, fmt.Errorf("unexpected array token: %q", token)
				}
			case ']':
				switch cur.Type() {
				case ast.Array:
					cur = cur.Parent()
				default:
					return nil, fmt.Errorf("unexpected array close token: %q", token)
				}
			default:
				return nil, fmt.Errorf("unexpected delimiter: %q", token)
			}
		default:
			return nil, fmt.Errorf("unexpected token type: %T", token)
		}
	}

	if cur != root {
		return nil, fmt.Errorf("unmatched block: %v", cur)
	}
	return root, nil
}
