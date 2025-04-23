package main

import (
	"context"

	// Packages
	server "github.com/mutablelogic/go-server"
	cert "github.com/mutablelogic/go-server/pkg/cert"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type task struct {
	*cert.CertManager
}

var _ server.Task = (*task)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newTaskWith(certmanager *cert.CertManager) *task {
	return &task{certmanager}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (t *task) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
