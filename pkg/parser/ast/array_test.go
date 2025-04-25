package ast_test

import (
	"testing"

	// Packages
	ast "github.com/mutablelogic/go-server/pkg/parser/ast"
	assert "github.com/stretchr/testify/assert"
)

func Test_Array(t *testing.T) {
	assert := assert.New(t)

	// Create a new array node
	node := ast.NewArray(nil)
	assert.NotNil(node)
	assert.Equal(ast.Array, node.Type())
	assert.Equal(nil, node.Parent())
	assert.Len(node.Children(), 0)

	// Append a child node
	ast.NewNull(node)
	assert.Len(node.Children(), 1)
	assert.Equal(node, node.Children()[0].Parent())
}
