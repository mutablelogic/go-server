package ast_test

import (
	"testing"

	// Packages
	ast "github.com/mutablelogic/go-server/pkg/provider/parser/ast"
	assert "github.com/stretchr/testify/assert"
)

func Test_Dict(t *testing.T) {
	assert := assert.New(t)

	// Create a new dict node
	node := ast.NewDict(nil)
	assert.NotNil(node)
	assert.Equal(ast.Dict, node.Type())
	assert.Equal(nil, node.Parent())
	assert.Len(node.Children(), 0)

	// Append a child node
	ast.NewIdent(node, "key")
	assert.Len(node.Children(), 1)
	assert.Equal(node, node.Children()[0].Parent())
}
