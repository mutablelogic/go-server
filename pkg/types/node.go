package types

import "fmt"

type NodeType uint

type Node struct {
	t NodeType
	v any
}

const (
	NodeString NodeType = iota
	NodeExpr
)

func NewString(v string) *Node {
	return &Node{t: NodeString, v: v}
}

func NewExpr(v string) *Node {
	return &Node{t: NodeExpr, v: v}
}

func (n *Node) String() string {
	switch n.t {
	case NodeString:
		return fmt.Sprintf("<str %q>", n.v.(string))
	case NodeExpr:
		return fmt.Sprintf("<expr %q>", n.v.(string))
	}
	panic("Unknown node type")
}
