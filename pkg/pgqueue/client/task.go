package client

import (
	"context"
	"net/http"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) CreateTask(ctx context.Context, queue string, task schema.TaskMeta) (*schema.Task, error) {
	type body struct {
		Queue string `json:"queue"`
		schema.TaskMeta
	}
	req, err := client.NewJSONRequest(body{
		Queue:    queue,
		TaskMeta: task,
	})
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.Task
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("task")); err != nil {
		return nil, err
	}

	// Return the response
	return &response, nil
}

func (c *Client) RetainTask(ctx context.Context, opts ...Opt) (*schema.Task, error) {
	// Apply options
	opt, err := applyOpts(opts...)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.Task
	if err := c.DoWithContext(ctx, nil, &response, client.OptPath("task"), client.OptQuery(opt.Values)); err != nil {
		return nil, err
	}

	// Return the response
	return &response, nil
}

func (c *Client) ReleaseTask(ctx context.Context, task uint64, err error) (*schema.TaskWithStatus, error) {
	var body struct {
		Result any `json:"result,omitempty"`
	}
	if err != nil {
		body.Result = err.Error()
	}
	req, err := client.NewJSONRequestEx(http.MethodPatch, body, "")
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.TaskWithStatus
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("task", task)); err != nil {
		return nil, err
	}

	// Return the response
	return &response, nil
}
