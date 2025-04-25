package ast

import (
	"encoding/json"
)

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
	node := &array{parent: parent, children: make([]Node, 0, 10)}
	if parent != nil {
		parent.AppendChild(node)
	}
	return node
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (node array) String() string {
	data, err := json.Marshal(node)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (node array) MarshalJSON() ([]byte, error) {
	return json.Marshal(node.children)
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
