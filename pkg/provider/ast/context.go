package ast

/////////////////////////////////////////////////////////////////////
// TYPES

// Evaluate a value node and return the value
type EvalFunc func(ctx *Context, value any) (any, error)

type Context struct {
	eval EvalFunc
}

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewContext(fn EvalFunc) *Context {
	return &Context{eval: fn}
}
