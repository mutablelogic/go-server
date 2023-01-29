package main

import (
	"context"
	"fmt"

	// Packages
	multierror "github.com/hashicorp/go-multierror"
	pool "github.com/mutablelogic/go-accessory/pkg/pool"
	iface "github.com/mutablelogic/go-server"
	task "github.com/mutablelogic/go-server/pkg/task"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-accessory"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// connection pool task
type connpool struct {
	task.Task
	Pool
}

var _ iface.Task = (*connpool)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new pool task from plugin configuration
func NewWithPlugin(ctx context.Context, p Plugin) (iface.Task, error) {
	this := new(connpool)

	// Get plugin options
	opts := p.Options()

	// TODO: trace

	// Get the URL
	if url, err := p.Url(); err != nil {
		return nil, err
	} else if pool := pool.New(ctx, url, opts...); pool == nil {
		return nil, ErrBadParameter.With(url)
	} else {
		this.Pool = pool
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (pool *connpool) String() string {
	str := "<" + defaultName
	if pool.Pool != nil {
		str += fmt.Sprint(" pool=", pool.Pool)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (pool *connpool) Run(ctx context.Context) error {
	var result error
	if pool.Pool == nil {
		return ErrInternalAppError.With("No pool")
	}

	// Wait for context to be cancelled
	<-ctx.Done()

	// close the connection pool
	if err := pool.Pool.Close(); err != nil {
		result = multierror.Append(result, err)
	}

	// Return any errors
	return result
}
