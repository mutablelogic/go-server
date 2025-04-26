package client

import (
	"context"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) CreateTask(ctx context.Context, queue string, task schema.TaskMeta) (*schema.Task, error) {
	req, err := client.NewJSONRequest(task)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.Task
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("task", queue)); err != nil {
		return nil, err
	}

	// Return the response
	return &response, nil
}
