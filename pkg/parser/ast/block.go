package ast

import "fmt"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type block struct {
	t        Type
	children []Node
}

var _ Node = (*block)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new block
func NewBlock() Node {
	return &block{t: Block}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (node block) String() string {
	str := "<" + fmt.Sprint(node.Type())
	for _, child := range node.children {
		str += " " + fmt.Sprint(child)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (node block) Type() Type {
	return node.t
}

func (node block) Children() []Node {
	return node.children
}

func (node block) Value() any {
	return nil
}

func (node *block) AppendChild(child Node) {
	node.children = append(node.children, child)
}
