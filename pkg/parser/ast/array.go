package ast

import "fmt"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type array struct {
	parent   Node
	children []Node
}

var _ Node = (*array)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new block
func NewArray(parent Node) Node {
	node := &array{parent: parent}
	if parent != nil {
		parent.AppendChild(node)
	}
	return node
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (node array) String() string {
	str := "<" + fmt.Sprint(node.Type())
	for _, child := range node.children {
		str += " " + fmt.Sprint(child)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (node array) Type() Type {
	return Array
}

func (node array) Parent() Node {
	return node.parent
}

func (node array) Children() []Node {
	return node.children
}

func (node array) Value() any {
	return nil
}

func (node *array) AppendChild(child Node) {
	node.children = append(node.children, child)
}
