package ast

import "encoding/json"

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
