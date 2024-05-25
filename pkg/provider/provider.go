package provider

import (
	"context"
	"errors"
	"log"
	"sync"

	// Packages
	server "github.com/mutablelogic/go-server"
)

////////////////////////////////////////////////////////////////////////////
// TYPES

// provider implements the server.Provider interface which runs a set of tasks
type provider struct {
	// All the tasks in order, with their current state
	tasks []state

	// All the loggers
	loggers []server.Logger
}

// state is the state of a task
type state struct {
	sync.WaitGroup
	server.Task

	Context context.Context
	Cancel  context.CancelFunc
	Label   string
}

var _ server.Task = (*provider)(nil)

////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName = "provider"
)

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new provider with tasks in the order they have been created.
func NewProvider(tasks ...server.Task) server.Provider {
	p := new(provider)

	// Enumerate all the tasks
	for _, task := range tasks {
		p.tasks = append(p.tasks, state{Task: task, Label: task.Label()})

		// If it's a logger, then add to the list of loggers
		if logger, ok := task.(server.Logger); ok {
			p.loggers = append(p.loggers, logger)
		}
	}

	// Return success
	return p
}

////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - TASK

// Run all the tasks in parallel, and cancel them all in the reverse order
// when the context is cancelled.
func (p *provider) Run(ctx context.Context) error {
	var result error

	// Run all the tasks in parallel
	for i := range p.tasks {
		// Create a context for each task
		ctx, cancel := context.WithCancel(WithLabel(context.Background(), p.tasks[i].Label))
		p.tasks[i].Context = ctx
		p.tasks[i].Cancel = cancel
		p.tasks[i].Add(1)

		// Run the task in a goroutine
		go func(i int) {
			defer p.tasks[i].Done()
			defer p.tasks[i].Cancel()

			p.Print(ctx, "Running")
			if err := p.tasks[i].Run(ctx); err != nil {
				p.Print(ctx, err)
				result = errors.Join(result, err)
			}
		}(i)
	}

	// Wait for the cancel
	<-ctx.Done()

	// Cancel all the tasks in reverse order, waiting for each to complete
	// before cancelling the next
	for i := len(p.tasks) - 1; i >= 0; i-- {
		p.Print(p.tasks[i].Context, "Stopping")
		p.tasks[i].Cancel()
		p.tasks[i].Wait()

		// TODO: If the task is in the set of loggers, then remove it
	}

	// Return any errors
	return result
}

func (p *provider) Label() string {
	return defaultName
}

////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - LOGGER

func (p *provider) Print(ctx context.Context, args ...any) {
	var wg sync.WaitGroup

	// With no loggers, just print to stdout
	if len(p.loggers) == 0 {
		log.Print(args...)
		return
	}

	// With one logger, just print to that logger
	if len(p.loggers) == 1 {
		p.loggers[0].Print(ctx, args...)
		return
	}

	// With more loggers, print to all of them in parallel
	for _, logger := range p.loggers {
		wg.Add(1)
		go func(logger server.Logger) {
			defer wg.Done()
			logger.Print(ctx, args...)
		}(logger)
	}
	wg.Wait()
}

func (p *provider) Printf(ctx context.Context, fmt string, args ...any) {
	var wg sync.WaitGroup

	// With no loggers, just print to stdout
	if len(p.loggers) == 0 {
		log.Printf(fmt, args...)
		return
	}

	// With one logger, just print to that logger
	if len(p.loggers) == 1 {
		p.loggers[0].Printf(ctx, fmt, args...)
		return
	}

	// With more loggers, print to all of them in parallel
	for _, logger := range p.loggers {
		wg.Add(1)
		go func(logger server.Logger) {
			defer wg.Done()
			logger.Printf(ctx, fmt, args...)
		}(logger)
	}
	wg.Wait()
}
