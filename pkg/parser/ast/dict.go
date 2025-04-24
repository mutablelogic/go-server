package ast

import "fmt"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type dict struct {
	parent   Node
	children []Node
}

var _ Node = (*dict)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new map[string]Node dictionary
func NewDict(parent Node) Node {
	node := &dict{parent: parent}
	if parent != nil {
		parent.AppendChild(node)
	}
	return node
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (node dict) String() string {
	str := "<" + fmt.Sprint(node.Type())
	for _, child := range node.children {
		str += " " + fmt.Sprint(child)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (node dict) Type() Type {
	return Dict
}

func (node dict) Parent() Node {
	return node.parent
}

func (node dict) Children() []Node {
	return node.children
}

func (node dict) Value() any {
	return nil
}

func (node *dict) AppendChild(child Node) {
	node.children = append(node.children, child)
}
