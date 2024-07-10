package ast

import (
	"encoding/json"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type fieldNode struct {
	N string
	C []Node
	P Node
}

var _ Node = (*fieldNode)(nil)

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewFieldNode(parent Node, name string) *fieldNode {
	return &fieldNode{
		P: parent,
		N: name,
	}
}

/////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r fieldNode) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (r fieldNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonNode{
		Type:     "Field",
		Name:     r.N,
		Children: r.C,
	})
}

/////////////////////////////////////////////////////////////////////
// METHODS

func (r *fieldNode) Type() NodeType {
	return Field
}

func (r *fieldNode) Parent() Node {
	return r.P
}

func (r *fieldNode) Children() []Node {
	return r.C
}

func (r *fieldNode) Append(n Node) Node {
	r.C = append(r.C, n)
	return n
}

func (r *fieldNode) Key() string {
	return r.N
}

func (r *fieldNode) Value(ctx *Context) (any, error) {
	if len(r.C) != 1 {
		return nil, ErrInternalAppError.With("FieldNode expected one child")
	}
	return r.C[0].Value(ctx)
}
