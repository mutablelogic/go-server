package ast

import "encoding/json"

/////////////////////////////////////////////////////////////////////
// TYPES

type mapNode struct {
	C []Node
	P Node
}

var _ Node = (*mapNode)(nil)

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewMapNode(parent Node) *mapNode {
	return &mapNode{
		P: parent,
	}
}

/////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r mapNode) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (r mapNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonNode{
		Type:     "Map",
		Children: r.C,
	})
}

/////////////////////////////////////////////////////////////////////
// METHODS

func (r *mapNode) Type() NodeType {
	return Map
}

func (r *mapNode) Parent() Node {
	return r.P
}

func (r *mapNode) Children() []Node {
	return r.C
}

func (r *mapNode) Append(n Node) Node {
	r.C = append(r.C, n)
	return n
}
