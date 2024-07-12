package ast

import (
	"strings"

	// Packages
	"github.com/mutablelogic/go-server/pkg/types"
)

/////////////////////////////////////////////////////////////////////
// TYPES

// Evaluate a value node and return the value
type EvalFunc func(ctx *Context, value any) (any, error)

type Context struct {
	label types.Label
	path  []string
	eval  EvalFunc
}

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewContext(fn EvalFunc) *Context {
	return &Context{eval: fn}
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get the label
func (ctx *Context) Label() types.Label {
	return ctx.label
}

// Set the label
func (ctx *Context) SetLabel(label types.Label) {
	ctx.label = label
}

// Return the current path
func (ctx *Context) Path() types.Label {
	return types.Label(strings.Join(ctx.path, ""))
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (ctx *Context) push(path string, sep bool) {
	if sep {
		ctx.path = append(ctx.path, types.LabelSeparator+path)
	} else {
		ctx.path = append(ctx.path, path)
	}
}

func (ctx *Context) pop() {
	ctx.path = ctx.path[:len(ctx.path)-1]
}
