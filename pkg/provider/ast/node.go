package ast

type NodeType int

const (
	_ NodeType = iota
	Root
	Plugin
	Map
	Array
	Field
	Value
)

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
	Children []Node `json:"children,omitempty"`
}

func (r NodeType) String() string {
	switch r {
	case Root:
		return "Root"
	case Plugin:
		return "Plugin"
	case Map:
		return "Map"
	case Array:
		return "Array"
	case Field:
		return "Field"
	case Value:
		return "Value"
	}
	return "Unknown"
}
