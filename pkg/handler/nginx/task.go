package nginx

import (
	"context"
	"errors"
	"os"
	"sync"
	"syscall"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Check interfaces are satisfied
var _ server.Task = (*nginx)(nil)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the label
func (task *nginx) Label() string {
	// TODO
	return defaultName
}

// Run the http server until the context is cancelled
func (task *nginx) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	var result error

	// TODO Remove the temporary directies (Config,Data) if they were created
	defer func() {
		var result error
		for _, path := range task.deletePaths {
			if _, err := os.Stat(path); err == nil {
				task.log(ctx, "Removing temporary path: "+path)
				if err := os.RemoveAll(path); err != nil {
					result = errors.Join(result, err)
				}
			}
		}
		if result != nil {
			task.log(ctx, result.Error())
		}
	}()

	// We need to copy the configuration files to the temporary directory
	// then reload the folder configuration
	if err := fsCopyTo(task.configPath); err != nil {
		return err
	} else if err := task.folders.Reload(); err != nil {
		return err
	}

	// Get the version of nginx
	if version, err := task.getVersion(); err != nil {
		return err
	} else {
		task.version = version
	}

	// Add stdout, stderr for the nginx commands
	fn := func(data []byte) {
		task.log(ctx, string(data))
	}
	task.run.Out, task.run.Err = fn, fn
	task.test.Out, task.test.Err = fn, fn

	// Run the nginx server in the background
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := task.run.Run(); err != nil {
			result = errors.Join(result, err)
		}
	}()

	// Runloop for nginx
	timer := time.NewTicker(time.Second)
	defer timer.Stop()
RUN_LOOP:
	for {
		select {
		case <-ctx.Done():
			break RUN_LOOP
		case <-timer.C:
			if task.run.Exited() {
				break RUN_LOOP
			}
		}
	}

	// Perform shutdown, escalating signals
	checkTicker := time.NewTicker(500 * time.Millisecond)
	signalTicker := time.NewTimer(100 * time.Millisecond)
	termSignal := syscall.SIGQUIT
	defer checkTicker.Stop()
	defer signalTicker.Stop()

FOR_LOOP:
	for {
		select {
		case <-checkTicker.C:
			if task.run.Exited() {
				break FOR_LOOP
			}
		case <-signalTicker.C:
			if !task.run.Exited() {
				if err := task.run.Signal(termSignal); err != nil {
					result = errors.Join(result, err)
				}
			}
			switch termSignal {
			case syscall.SIGQUIT:
				task.log(ctx, "Sending graceful shutdown signal")
				termSignal = syscall.SIGTERM
				signalTicker.Reset(20 * time.Second) // Escalate after 20 seconds
			case syscall.SIGTERM:
				task.log(ctx, "Sending immediate shutdown signal")
				termSignal = syscall.SIGKILL
				signalTicker.Reset(5 * time.Second) // Escalate after 5 seconds
			case syscall.SIGKILL:
				task.log(ctx, "Sending kill signal")
				signalTicker.Reset(5 * time.Second) // Continue sending every 5 seconds
			}
		}
	}

	// Wait for nginx to exit
	wg.Wait()

	// Return any errors
	return result
}
