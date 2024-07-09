package ast

import "encoding/json"

/////////////////////////////////////////////////////////////////////
// TYPES

type arrayNode struct {
	C []Node
	P Node
}

var _ Node = (*arrayNode)(nil)

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewArrayNode(parent Node) *arrayNode {
	return &arrayNode{
		P: parent,
	}
}

/////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r arrayNode) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (r arrayNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonNode{
		Type:     "Array",
		Children: r.C,
	})
}

/////////////////////////////////////////////////////////////////////
// METHODS

func (r *arrayNode) Type() NodeType {
	return Array
}

func (r *arrayNode) Parent() Node {
	return r.P
}

func (r *arrayNode) Children() []Node {
	return r.C
}

func (r *arrayNode) Append(n Node) Node {
	r.C = append(r.C, n)
	return n
}
