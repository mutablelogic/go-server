package config

import (
	"context"

	// Packages
	pg "github.com/djthorpe/go-pg"
	server "github.com/mutablelogic/go-server"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type pgpool struct {
	conn pg.PoolConn
}

var _ server.Task = (*pgpool)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewTask(conn pg.PoolConn) *pgpool {
	return &pgpool{conn}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (pg *pgpool) Conn() pg.PoolConn {
	return pg.conn
}

func (*pgpool) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
