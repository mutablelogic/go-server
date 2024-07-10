package ast

import (
	"encoding/json"
	"fmt"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type valueNode struct {
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
	if len(r.C) == 0 {
		return json.Marshal(r.V)
	}
	return json.Marshal(jsonNode{
		Type:     "MapEntry",
		Name:     fmt.Sprint(r.V),
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
	if len(r.C) > 0 {
		return fmt.Sprint(r.V)
	} else {
		return ""
	}
}

func (r *valueNode) Value(ctx *Context) (any, error) {
	if ctx.eval == nil {
		return r.V, nil
	}
	if len(r.C) == 0 {
		return ctx.eval(ctx, r)
	} else {
		ctx.push(r.Key())
		defer ctx.pop()
		return ctx.eval(ctx, r)
	}
}
