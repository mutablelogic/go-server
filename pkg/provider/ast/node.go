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
