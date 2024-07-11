package ast

import (
	"encoding/json"
)

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
// PUBLIC METHODS

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

func (r *mapNode) Key() string {
	return ""
}

func (r *mapNode) Value(ctx *Context) (any, error) {
	var err error
	result := make(map[string]any, len(r.C))
	for _, child := range r.C {
		key := child.Key()

		ctx.push(key)
		value, err := child.Children()[0].Value(ctx)
		ctx.pop()
		if err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, err
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (r *mapNode) containsKey(key string) bool {
	for _, child := range r.C {
		if child.Key() == key {
			return true
		}
	}
	return false
}
