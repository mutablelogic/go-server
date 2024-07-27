package ast

/////////////////////////////////////////////////////////////////
// TYPES

type NodeType int

type Node interface {
	// Type of node
	Type() NodeType

	// Parent node
	Parent() Node

	// Node key (plugin name, map key)
	Key() string

	// Node evaluated value
	Value(ctx *Context) (any, error)

	// Child nodes
	Children() []Node

	// Add child node, and return it
	Append(Node) Node
}

type jsonNode struct {
	Type     string `json:"type"`
	Name     string `json:"name,omitempty"`
	Value    any    `json:"value,omitempty"`
	Children []Node `json:"children,omitempty"`
}

/////////////////////////////////////////////////////////////////
// GLOBALS

const (
	_ NodeType = iota
	Map
	Array
	Value
)

/////////////////////////////////////////////////////////////////
// STRINGIFY

func (r NodeType) String() string {
	switch r {
	case Map:
		return "Map"
	case Array:
		return "Array"
	case Value:
		return "Value"
	}
	return "Unknown"
}
