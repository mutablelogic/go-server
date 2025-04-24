package ast

import (
	"fmt"
	"strconv"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type value struct {
	p Node
	t Type
	s string
	b bool
	c []Node
}

var _ Node = (*value)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewString(parent Node, v string) Node {
	node := &value{p: parent, t: String, s: v}
	if parent != nil {
		parent.AppendChild(node)
	}
	return node
}

func NewIdent(parent Node, v string) Node {
	node := &value{p: parent, t: Ident, s: v}
	if parent != nil {
		parent.AppendChild(node)
	}
	return node
}

func NewBool(parent Node, v bool) Node {
	node := &value{p: parent, t: Bool, b: v}
	if parent != nil {
		parent.AppendChild(node)
	}
	return node
}

func NewNumber(parent Node, v string) Node {
	node := &value{p: parent, t: Number, s: v}
	if parent != nil {
		parent.AppendChild(node)
	}
	return node
}

func NewNull(parent Node) Node {
	node := &value{p: parent, t: Null}
	if parent != nil {
		parent.AppendChild(node)
	}
	return node
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (node value) String() string {
	str := "<" + fmt.Sprint(node.Type())
	switch node.t {
	case Ident:
		str += " " + fmt.Sprintf("%q", node.s)
	case String:
		str += " " + fmt.Sprintf("%q", node.s)
	case Number:
		str += " " + fmt.Sprint(node.s)
	case Bool:
		str += " " + fmt.Sprint(node.b)
	}
	for _, child := range node.c {
		str += " " + fmt.Sprint(child)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (node value) Type() Type {
	return node.t
}

func (node value) Parent() Node {
	return node.p
}

func (node value) Children() []Node {
	return node.c
}

func (node value) Value() any {
	switch node.t {
	case String:
		return node.s
	case Bool:
		return node.b
	case Number:
		if f, err := strconv.ParseFloat(node.s, 64); err != nil {
			return nil
		} else {
			return f
		}
	}
	return nil
}

func (node *value) AppendChild(child Node) {
	node.c = append(node.c, child)
}
