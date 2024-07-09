package ast

import "encoding/json"

/////////////////////////////////////////////////////////////////////
// TYPES

type valueNode struct {
	V any
	P Node
}

var _ Node = (*valueNode)(nil)

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewValueNode(parent Node, value any) *valueNode {
	return &valueNode{
		P: parent,
		V: value,
	}
}

/////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r valueNode) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (r valueNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.V)
}

/////////////////////////////////////////////////////////////////////
// METHODS

func (r *valueNode) Type() NodeType {
	return Value
}

func (r *valueNode) Parent() Node {
	return r.P
}

func (r *valueNode) Children() []Node {
	// No children
	return nil
}

func (r *valueNode) Append(n Node) Node {
	// No children
	return nil
}
