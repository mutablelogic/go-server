package ast

import (
	"encoding/json"
	"errors"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type mapNode struct {
	C []Node
	P Node
}

var _ Node = (*mapNode)(nil)

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewMapNode(parent Node) *mapNode {
	return &mapNode{
		P: parent,
	}
}

/////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r mapNode) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (r mapNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonNode{
		Type:     "Map",
		Children: r.C,
	})
}

/////////////////////////////////////////////////////////////////////
// METHODS

func (r *mapNode) Type() NodeType {
	return Map
}

func (r *mapNode) Parent() Node {
	return r.P
}

func (r *mapNode) Children() []Node {
	return r.C
}

func (r *mapNode) Append(n Node) Node {
	r.C = append(r.C, n)
	return n
}

func (r *mapNode) Key() string {
	return ""
}

func (r *mapNode) Value(ctx *Context) (any, error) {
	var err error
	result := make(map[string]any, len(r.C))
	for _, child := range r.C {
		key, err_ := child.Value(nil)
		value := child.Children()
		if err_ != nil {
			err = errors.Join(err, err_)
			continue
		}

		keyStr, ok := key.(string)
		if !ok {
			err = errors.Join(err, ErrInternalAppError.With("expected string key"))
			continue
		}
		if len(value) != 1 {
			err = errors.Join(err, ErrInternalAppError.With("FieldNode expected one child"))
			continue
		}
		if _, exists := result[keyStr]; exists {
			err = errors.Join(err, ErrDuplicateEntry.Withf("%q", keyStr))
			continue
		}

		ctx.push(keyStr)
		value_, err_ := value[0].Value(ctx)
		ctx.pop()

		if err_ != nil {
			err = errors.Join(err, err_)
			continue
		}
		result[keyStr] = value_
	}
	return result, err
}
