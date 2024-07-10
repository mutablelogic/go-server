package ast

import "strings"

/////////////////////////////////////////////////////////////////////
// TYPES

// Evaluate a value node and return the value
type EvalFunc func(ctx *Context, value any) (any, error)

type Context struct {
	path []string
	eval EvalFunc
}

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewContext(fn EvalFunc) *Context {
	return &Context{eval: fn}
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the current path
func (ctx *Context) Path() string {
	return "/" + strings.Join(ctx.path, "/")
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (ctx *Context) push(path string) {
	ctx.path = append(ctx.path, path)
}

func (ctx *Context) pop() {
	ctx.path = ctx.path[:len(ctx.path)-1]
}
