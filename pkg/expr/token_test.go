package expr_test

import (
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-server/pkg/expr"
)

func Test_Token_001(t *testing.T) {
	assert := assert.New(t)
	t.Run("Any", func(t *testing.T) {
		tok := NewToken(Any, "hello", Pos{Path: nil, Line: 1, Col: 1})
		assert.NotNil(tok)
		assert.Equal(Any, tok.Kind)
		assert.Equal("hello", tok.Val)
		assert.Equal("pos<@1:1>", tok.Pos.String())
	})
	t.Run("String", func(t *testing.T) {
		tok := NewToken(String, "hello", Pos{Path: nil, Line: 1, Col: 1})
		assert.NotNil(tok)
		assert.Equal(String, tok.Kind)
		assert.Equal("hello", tok.Val)
		assert.Equal("pos<@1:1>", tok.Pos.String())
	})
	t.Run("Expr", func(t *testing.T) {
		tok := NewToken(Expr, "hello", Pos{Path: nil, Line: 1, Col: 1})
		assert.NotNil(tok)
		assert.Equal(Expr, tok.Kind)
		assert.Equal("hello", tok.Val)
		assert.Equal("pos<@1:1>", tok.Pos.String())
	})
}
