package ast

import (
	"fmt"
	"strconv"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type value struct {
	t Type
	v string
	b bool
}

var _ Node = (*value)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewString(v string) Node {
	return &value{t: String, v: v}
}

func NewBool(v bool) Node {
	return &value{t: Bool, b: v}
}

func NewNumber(v string) Node {
	return &value{t: Number, v: v}
}

func NewNull() Node {
	return &value{t: Null}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (node value) String() string {
	str := "<" + fmt.Sprint(node.Type())
	switch node.t {
	case String:
		str += " " + fmt.Sprintf("%q", node.v)
	case Number:
		str += " " + fmt.Sprint(node.v)
	case Bool:
		str += " " + fmt.Sprint(node.b)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (node value) Type() Type {
	return node.t
}

func (value) Children() []Node {
	return nil
}

func (node value) Value() any {
	switch node.t {
	case String:
		return node.v
	case Bool:
		return node.b
	case Number:
		if f, err := strconv.ParseFloat(node.v, 64); err != nil {
			return nil
		} else {
			return f
		}
	}
	return nil
}

func (value) AppendChild(Node) {
	// Does nothing
}
