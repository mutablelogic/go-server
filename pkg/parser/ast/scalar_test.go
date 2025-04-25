package ast_test

import (
	"testing"

	// Packages
	ast "github.com/mutablelogic/go-server/pkg/parser/ast"
	assert "github.com/stretchr/testify/assert"
)

func Test_Bool_001(t *testing.T) {
	assert := assert.New(t)

	// Create a new dict node
	node := ast.NewBool(nil, true)
	assert.NotNil(node)
	assert.Equal(ast.Bool, node.Type())
	assert.Equal(nil, node.Parent())
	assert.Len(node.Children(), 0)
	assert.Equal(true, node.Value())
}

func Test_Bool_002(t *testing.T) {
	assert := assert.New(t)

	// Create a new dict node
	node := ast.NewBool(nil, false)
	assert.NotNil(node)
	assert.Equal(ast.Bool, node.Type())
	assert.Equal(nil, node.Parent())
	assert.Len(node.Children(), 0)
	assert.Equal(false, node.Value())
}

func Test_String(t *testing.T) {
	assert := assert.New(t)

	// Create a new dict node
	node := ast.NewString(nil, "test")
	assert.NotNil(node)
	assert.Equal(ast.String, node.Type())
	assert.Equal(nil, node.Parent())
	assert.Len(node.Children(), 0)
	assert.Equal("test", node.Value())
}

func Test_Ident(t *testing.T) {
	assert := assert.New(t)

	// Create a new dict node
	node := ast.NewIdent(nil, "test")
	assert.NotNil(node)
	assert.Equal(ast.Ident, node.Type())
	assert.Equal(nil, node.Parent())
	assert.Len(node.Children(), 0)
	assert.Equal("test", node.Value())
}

func Test_Null(t *testing.T) {
	assert := assert.New(t)

	// Create a new dict node
	node := ast.NewNull(nil)
	assert.NotNil(node)
	assert.Equal(ast.Null, node.Type())
	assert.Equal(nil, node.Parent())
	assert.Len(node.Children(), 0)
	assert.Equal(nil, node.Value())
}

func Test_Number_001(t *testing.T) {
	assert := assert.New(t)

	// Create a new dict node
	node := ast.NewNumber(nil, "123")
	assert.NotNil(node)
	assert.Equal(ast.Number, node.Type())
	assert.Equal(nil, node.Parent())
	assert.Len(node.Children(), 0)
	assert.Equal(float64(123), node.Value())
}

func Test_Number_002(t *testing.T) {
	assert := assert.New(t)

	// Create a new dict node
	node := ast.NewNumber(nil, "ABC")
	assert.NotNil(node)
	assert.Equal(ast.Number, node.Type())
	assert.Equal(nil, node.Parent())
	assert.Len(node.Children(), 0)
	assert.Equal(nil, node.Value())
}
