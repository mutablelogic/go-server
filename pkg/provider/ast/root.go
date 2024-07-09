package ast

import "encoding/json"

/////////////////////////////////////////////////////////////////////
// TYPES

type rootNode struct {
	C []Node
}

var _ Node = (*rootNode)(nil)

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewRootNode() *rootNode {
	return &rootNode{}
}

/////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r rootNode) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (r rootNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonNode{
		Type:     "Root",
		Children: r.C,
	})
}

/////////////////////////////////////////////////////////////////////
// METHODS

func (r *rootNode) Type() NodeType {
	return Root
}

func (r *rootNode) Parent() Node {
	return nil
}

func (r *rootNode) Children() []Node {
	return r.C
}

func (r *rootNode) Append(n Node) Node {
	r.C = append(r.C, n)
	return n
}
