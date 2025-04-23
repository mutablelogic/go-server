package jsonparser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/mutablelogic/go-server/pkg/parser/ast"
)

func Read(r io.Reader) error {
	parser := json.NewDecoder(r)
	parser.UseNumber() // Use Number to preserve number types
	for {
		token, err := parser.Token()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}
		switch token := token.(type) {
		case nil:
			node := ast.NewNull()
			fmt.Println(node)
		case string:
			node := ast.NewString(token)
			fmt.Println(node)
		case bool:
			node := ast.NewBool(token)
			fmt.Println(node)
		case json.Number:
			node := ast.NewNumber(token.String())
			fmt.Println(node)
		case json.Delim:
			fmt.Printf("delim=%s\n", token)
		default:
			return fmt.Errorf("unexpected token type: %T", token)
		}
	}
	return nil
}
