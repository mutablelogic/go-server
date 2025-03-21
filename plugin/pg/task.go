package pg

import (
	"context"

	// Packages
	pg "github.com/djthorpe/go-pg"
	server "github.com/mutablelogic/go-server"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type pgpool struct {
	pg.PoolConn
}

var _ server.Task = (*pgpool)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func taskWithConn(conn pg.PoolConn) *pgpool {
	return &pgpool{conn}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (*pgpool) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
