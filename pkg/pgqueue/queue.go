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
// PUBLIC METHODS

// RegisterQueue creates a new queue, or updates an existing queue, and returns it.
func (client *Client) RegisterQueue(ctx context.Context, meta schema.Queue) (*schema.Queue, error) {
	var queue schema.Queue
	if err := client.conn.Tx(ctx, func(conn pg.Conn) error {
		// Get a queue
		if err := conn.Get(ctx, &queue, schema.QueueName(meta.Queue)); err != nil && !errors.Is(err, pg.ErrNotFound) {
			return err
		} else if errors.Is(err, pg.ErrNotFound) {
			// If the queue does not exist, then create it
			return conn.Insert(ctx, &queue, meta)
		} else {
			// If the queue exists, then update it
			return conn.Update(ctx, &queue, schema.QueueName(meta.Queue), meta)
		}
	}); err != nil {
		return nil, err
	}
	return &queue, nil
}

// CreateQueue creates a new queue, and returns it.
func (client *Client) CreateQueue(ctx context.Context, meta schema.Queue) (*schema.Queue, error) {
	var queue schema.Queue
	if err := client.conn.Tx(ctx, func(conn pg.Conn) error {
		if err := client.conn.Insert(ctx, &queue, meta); err != nil {
			return err
		} else if err := conn.Update(ctx, &queue, schema.QueueName(queue.Queue), meta); err != nil {
			return err
		}
		// Commit the transaction
		return nil
	}); err != nil {
		return nil, err
	}
	return &queue, nil
}

// GetQueue returns a queue with the given name.
func (client *Client) GetQueue(ctx context.Context, name string) (*schema.Queue, error) {
	var queue schema.Queue
	if err := client.conn.Get(ctx, &queue, schema.QueueName(name)); err != nil {
		if errors.Is(err, pg.ErrNotFound) {
			return nil, httpresponse.ErrNotFound.Withf("Queue %q not found", name)
		}
		return nil, err
	}
	return &queue, nil
}

// DeleteQueue deletes a queue with the given name, and returns the deleted queue.
func (client *Client) DeleteQueue(ctx context.Context, name string) (*schema.Queue, error) {
	var queue schema.Queue
	if err := client.conn.Delete(ctx, &queue, schema.QueueName(name)); err != nil {
		if errors.Is(err, pg.ErrNotFound) {
			return nil, httpresponse.ErrNotFound.Withf("Queue %q not found", name)
		}
		return nil, err
	}
	return &queue, nil
}

// UpdateQueue updates an existing queue with the given name, and returns the queue.
func (client *Client) UpdateQueue(ctx context.Context, name string, meta schema.Queue) (*schema.Queue, error) {
	var queue schema.Queue
	if err := client.conn.Update(ctx, &queue, schema.QueueName(name), meta); err != nil {
		if errors.Is(err, pg.ErrNotFound) {
			return nil, httpresponse.ErrNotFound.Withf("Queue %q not found", name)
		}
		return nil, err
	}
	return &queue, nil
}

// ListQueues returns all queues as a list
func (client *Client) ListQueues(ctx context.Context, req schema.QueueListRequest) (*schema.QueueList, error) {
	var list schema.QueueList

	// Perform list
	list.Body = make([]schema.Queue, 0, 10)
	if err := client.conn.List(ctx, &list, req); err != nil {
		return nil, err
	}
	return &list, nil
}
