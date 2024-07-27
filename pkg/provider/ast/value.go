package ast

import (
	"encoding/json"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type valueNode struct {
	K string
	V any
	P Node
	C []Node
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

func NewMapValueNode(parent Node, key string) (*valueNode, error) {
	if parent.(*mapNode).containsKey(key) {
		return nil, ErrDuplicateEntry.Withf("%q", key)
	}
	return &valueNode{
		P: parent,
		K: key,
	}, nil
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
	return json.Marshal(jsonNode{
		Type:     "Value",
		Name:     r.K,
		Value:    r.V,
		Children: r.C,
	})
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
	return r.C
}

func (r *valueNode) Append(n Node) Node {
	r.C = append(r.C, n)
	return n
}

func (r *valueNode) Key() string {
	return r.K
}

func (r *valueNode) Value(ctx *Context) (any, error) {
	if ctx == nil || ctx.eval == nil {
		return nil, ErrInternalAppError.With("Missing context evaluation function")
	}
	if len(r.C) > 0 {
		return r.C[0].Value(ctx)
	}

	// Evaluate the value
	return ctx.eval(ctx, r.V)
}
