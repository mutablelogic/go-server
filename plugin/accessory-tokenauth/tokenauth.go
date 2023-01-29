package main

import (
	// Packages
	"context"
	"errors"
	"fmt"

	"github.com/mutablelogic/go-accessory/pkg/auth"
	iface "github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/event"
	task "github.com/mutablelogic/go-server/pkg/task"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-accessory"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type tokenauth struct {
	task.Task
	Auth

	// The label of the task
	label string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new connection pool task from plugin configuration
func NewWithPlugin(plugin Plugin, label string) (iface.Task, error) {
	this := new(tokenauth)

	if pool := plugin.Pool(); pool == nil {
		return nil, ErrBadParameter.With("pool")
	} else {
		this.Auth = auth.New(pool.(Pool))
		this.label = label
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINIGFY

func (tokenauth *tokenauth) String() string {
	str := "<tokenauth"
	if tokenauth.label != "" {
		str += fmt.Sprintf(" label=%q", tokenauth.label)
	}
	if tokenauth.Auth != nil {
		str += fmt.Sprint(" auth=", tokenauth.Auth)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (tokenauth *tokenauth) Run(ctx context.Context) error {
	// Create an admin user if one doesn't exist
	if err := tokenauth.createAdmin(ctx); err != nil {
		return err
	}

	// Wait for context to be cancelled
	<-ctx.Done()

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (tokenauth *tokenauth) createAdmin(ctx context.Context) error {
	// Check for existing valid admin user and all scopes
	for _, scope := range adminScopes {
		err := tokenauth.Valid(ctx, defaultAdmin, scope)
		if err == nil {
			return nil
		} else if !errors.Is(err, ErrNotFound) {
			return fmt.Errorf("admin user is mis-configured or other error: %w", err)
		}
	}

	// Create the admin user with no expiry and scopes. Emit an event on creation.
	if value, err := tokenauth.CreateByte16(ctx, defaultAdmin, 0, adminScopes...); err != nil {
		return err
	} else {
		tokenauth.Emit(event.New(ctx, EventCreateAdmin, value))
	}

	// Return success
	return nil
}
