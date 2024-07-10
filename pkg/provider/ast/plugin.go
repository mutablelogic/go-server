package ast

import (
	"encoding/json"
	"errors"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type pluginNode struct {
	N string
	P Node
	C []Node
}

var _ Node = (*pluginNode)(nil)

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewPluginNode(parent Node, name string) *pluginNode {
	return &pluginNode{
		N: name,
		P: parent,
	}
}

/////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r pluginNode) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (r pluginNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonNode{
		Type:     "Plugin",
		Name:     r.N,
		Children: r.C,
	})
}

/////////////////////////////////////////////////////////////////////
// METHODS

func (r *pluginNode) Type() NodeType {
	return Plugin
}

func (r *pluginNode) Parent() Node {
	return r.P
}

func (r *pluginNode) Children() []Node {
	return r.C
}

func (r *pluginNode) Append(n Node) Node {
	r.C = append(r.C, n)
	return n
}

func (r *pluginNode) Key() string {
	return r.N
}

func (r *pluginNode) Value(ctx *Context) (any, error) {
	var err error
	result := make([]any, len(r.C))
	for i, child := range r.C {
		value, err_ := child.Value(ctx)
		if err != nil {
			err = errors.Join(err, err_)
		}
		result[i] = value
	}
	return result, err
}
