package schema

import (
	"context"
	"time"

	// Packages
	pg "github.com/djthorpe/go-pg"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	SchemaName       = "pgqueue"
	DefaultNamespace = "default"
	DefaultPrefix    = "/queue/v1"
	TopicQueueInsert = "queue_insert"
	QueueListLimit   = 100
	TaskListLimit    = 100
	TickerPeriod     = 15 * time.Second
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Bootstrap(ctx context.Context, conn pg.Conn) error {
	// Create the schema
	if err := pg.SchemaCreate(ctx, conn, SchemaName); err != nil {
		return err
	}
	// Create types, tables, ...
	if err := bootstrapQueue(ctx, conn); err != nil {
		return err
	}
	if err := bootstrapTask(ctx, conn); err != nil {
		return err
	}
	if err := bootstrapTicker(ctx, conn); err != nil {
		return err
	}
	// Commit the transaction
	return nil

}
