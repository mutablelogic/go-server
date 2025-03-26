package pgqueue

import (
	"context"
	"errors"

	// Packages
	pg "github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	conn     pg.PoolConn
	listener pg.Listener
	worker   string
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(ctx context.Context, conn pg.PoolConn, opt ...Opt) (*Client, error) {
	self := new(Client)
	opts, err := applyOpts(opt...)
	if err != nil {
		return nil, err
	} else {
		self.worker = opts.worker
	}

	// Create a listener
	if listener := pg.NewListener(conn); listener == nil {
		return nil, httpresponse.ErrInternalError.Withf("Cannot create listener")
	} else {
		self.listener = listener
	}

	// Set the connection
	self.conn = conn.With(
		"schema", schema.SchemaName,
		"ns", opts.namespace,
	).(pg.PoolConn)

	// Listen for topics
	for _, topic := range []string{opts.namespace + "_" + schema.TopicQueueInsert} {
		if err := self.listener.Listen(ctx, topic); err != nil {
			return nil, httpresponse.ErrInternalError.Withf("Cannot listen to topic %q: %v", topic, err)
		}
	}

	// If the schema does not exist, then bootstrap it
	if err := self.conn.Tx(ctx, func(conn pg.Conn) error {
		if exists, err := pg.SchemaExists(ctx, conn, schema.SchemaName); err != nil {
			return err
		} else if !exists {
			return schema.Bootstrap(ctx, conn)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Return success
	return self, nil
}

func (client *Client) Close(ctx context.Context) error {
	var result error
	if client.listener != nil {
		result = errors.Join(result, client.listener.Close(ctx))
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Worker returns the worker name
func (client *Client) Worker() string {
	return client.worker
}
