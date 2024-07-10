package ast

import (
	"encoding/json"
	"errors"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type rootNode struct {
	C []Node
}

var _ Node = (*rootNode)(nil)

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewRootNode() *rootNode {
	return &rootNode{}
}

/////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r rootNode) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (r rootNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonNode{
		Type:     "Root",
		Children: r.C,
	})
}

/////////////////////////////////////////////////////////////////////
// METHODS

func (r *rootNode) Type() NodeType {
	return Root
}

func (r *rootNode) Parent() Node {
	return nil
}

func (r *rootNode) Children() []Node {
	return r.C
}

func (r *rootNode) Append(n Node) Node {
	r.C = append(r.C, n)
	return n
}

func (r *rootNode) Key() string {
	return ""
}

func (r *rootNode) Value(ctx *Context) (any, error) {
	var err error
	result := make(map[string]any, len(r.C))
	for _, child := range r.C {
		key := child.Key()
		value, err_ := child.Value(ctx)
		if err != nil {
			err = errors.Join(err, err_)
		} else if _, exists := result[key]; exists {
			err = errors.Join(err, ErrDuplicateEntry.Withf("%q", key))
		} else {
			result[key] = value
		}
	}
	return result, err
}
