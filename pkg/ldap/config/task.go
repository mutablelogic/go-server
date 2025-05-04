package config

import (
	"context"

	// Packages
	server "github.com/mutablelogic/go-server"
	ldap "github.com/mutablelogic/go-server/pkg/ldap"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type task struct {
	manager *ldap.Manager
}

var _ server.Task = (*task)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewTask(manager *ldap.Manager) (server.Task, error) {
	self := new(task)
	self.manager = manager
	return self, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (task *task) Run(ctx context.Context) error {
	return task.manager.Run(ctx)
}
