package expr

import (
	"fmt"
	"strings"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Node is a node in the graph
type Node interface {
	Kind() TokenKind
}

///////////////////////////////////////////////////////////////////////////////
// EXPRESSIONS

type ExpressionListNode struct {
	v []Node
}

func (n *ExpressionListNode) Kind() TokenKind {
	return List
}

func (n *ExpressionListNode) String() string {
	result := make([]string, 0, len(n.v))
	for _, node := range n.v {
		result = append(result, fmt.Sprint(node))
	}
	return strings.Join(result, ",")
}

///////////////////////////////////////////////////////////////////////////////
// UNSIGNED INTEGER

// UintNumberNode is a node representing a number in unsigned uint64 format
type UintNumberNode struct {
	v uint64
}

func (n *UintNumberNode) Kind() TokenKind {
	return Number
}

func (n *UintNumberNode) String() string {
	return fmt.Sprint(n.v)
}

///////////////////////////////////////////////////////////////////////////////
// SIGNED INTEGER

// IntNumberNode is a node representing a number in signed uint64 format
type IntNumberNode struct {
	v int64
}

func (n *IntNumberNode) Kind() TokenKind {
	return Number
}

func (n *IntNumberNode) String() string {
	return fmt.Sprint(n.v)
}

///////////////////////////////////////////////////////////////////////////////
// FLOAT

// FloatNumberNode is a node representing a number in unsigned float64 format
type FloatNumberNode struct {
	v float64
}

func (n *FloatNumberNode) Kind() TokenKind {
	return Number
}

func (n *FloatNumberNode) String() string {
	return fmt.Sprint(n.v)
}

///////////////////////////////////////////////////////////////////////////////
// UNARY FUNCTION

type UnaryFunction struct {
	k TokenKind
	v Node
}

func (n *UnaryFunction) Kind() TokenKind {
	return n.k
}

func (n *UnaryFunction) String() string {
	return fmt.Sprintf("%s(%v)", n.k, n.v)
}

///////////////////////////////////////////////////////////////////////////////
// BINARY FUNCTION

type BinaryFunction struct {
	k    TokenKind
	l, r Node
}

func (n *BinaryFunction) Kind() TokenKind {
	return n.k
}

func (n *BinaryFunction) String() string {
	return fmt.Sprintf("%s(%v,%v)", n.k, n.l, n.r)
}
