package ast

import (
	"bytes"
	"encoding/json"
	"errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type dict struct {
	parent   Node
	children []Node
}

var _ Node = (*dict)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new map[string]Node dictionary
func NewDict(parent Node) Node {
	node := &dict{parent: parent}
	if parent != nil {
		parent.AppendChild(node)
	}
	return node
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (node dict) String() string {
	data, err := json.Marshal(node)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (node dict) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	// Write the JSON in the same order that the children were added
	buf.WriteRune('{')
	enc := json.NewEncoder(&buf)
	for i, child := range node.children {
		if child.Type() != Ident || len(child.Children()) != 1 {
			return nil, errors.New("invalid child type")
		}

		key, ok := child.Value().(string)
		if !ok {
			return nil, errors.New("invalid child type")
		}
		if i > 0 {
			buf.WriteRune(',')
		}
		enc.Encode(key)
		buf.WriteRune(':')
		enc.Encode(child.Children()[0])
	}
	buf.WriteRune('}')
	return buf.Bytes(), nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (node dict) Type() Type {
	return Dict
}

func (node dict) Parent() Node {
	return node.parent
}

func (node dict) Children() []Node {
	return node.children
}

func (node dict) Value() any {
	value := make(map[string]Node, len(node.children))
	for _, child := range node.children {
		if child.Type() != Ident || len(child.Children()) != 1 {
			return nil
		} else if key, ok := child.Value().(string); !ok {
			return nil
		} else {
			value[key] = child.Children()[0]
		}
	}
	return value
}

func (node *dict) AppendChild(child Node) {
	node.children = append(node.children, child)
}
