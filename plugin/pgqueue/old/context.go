package main

import (
	"context"

	// Packages
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
)

//////////////////////////////////////////////////////////////////////////////////
// TYPES

type ctxKey int

//////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ctxKeyTicker ctxKey = iota
	ctxKeyTask
)

//////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Ticker(ctx context.Context) *schema.Ticker {
	if v, ok := ctx.Value(ctxKeyTicker).(*schema.Ticker); ok {
		return v
	}
	return nil
}

func Task(ctx context.Context) *schema.Task {
	if v, ok := ctx.Value(ctxKeyTask).(*schema.Task); ok {
		return v
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func contextWithTicker(ctx context.Context, ticker *schema.Ticker) context.Context {
	return context.WithValue(ctx, ctxKeyTicker, ticker)
}

func contextWithTask(ctx context.Context, ticker *schema.Task) context.Context {
	return context.WithValue(ctx, ctxKeyTask, ticker)
}
