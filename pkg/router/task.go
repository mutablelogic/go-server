package router

import (
	"net/http"

	// Package imports
	task "github.com/mutablelogic/go-server/pkg/task"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type t struct {
	task.Task
	http.Handler
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func NewWithPlugin(p Plugin) (*t, error) {
	this := new(t)
	this.Handler = http.NewServeMux()
	return this, nil
}
