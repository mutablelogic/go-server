package pg

import (
	"context"

	// Packages
	pg "github.com/djthorpe/go-pg"
	server "github.com/mutablelogic/go-server"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type authtask struct {
	conn pg.PoolConn
}

var _ server.Task = (*authtask)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewTask(conn pg.PoolConn) *authtask {
	return &authtask{conn}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (pg *authtask) Conn() pg.PoolConn {
	return pg.conn
}

func (*authtask) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
