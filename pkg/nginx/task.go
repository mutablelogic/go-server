package nginx

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"syscall"
	"time"

	// Package imports
	multierror "github.com/hashicorp/go-multierror"
	event "github.com/mutablelogic/go-server/pkg/event"
	task "github.com/mutablelogic/go-server/pkg/task"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type state uint

type t struct {
	task.Task
	cfg     *Config
	cmd     *Cmd
	test    *Cmd
	version []byte
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	stateNone state = iota
	stateRun
	stateStop
	stateTerm
	stateKill
	stateInfo
	stateErr
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func NewWithPlugin(p Plugin) (*t, error) {
	this := new(t)

	// Create the command
	if cmd, err := NewWithCommand(p.Path(), p.Flags()...); err != nil {
		return nil, err
	} else if test, err := NewWithCommand(p.Path(), p.Flags()...); err != nil {
		return nil, err
	} else {
		this.cmd = cmd
		this.test = test
		this.test.cmd.Args = append(this.test.cmd.Args, "-t", "-q")
	}

	// Add the environment variables
	if err := this.cmd.SetEnv(p.Env()); err != nil {
		return nil, err
	} else if err := this.test.SetEnv(p.Env()); err != nil {
		return nil, err
	}

	// Add the available configurations
	if cfg, err := NewConfig(p.Available(), p.Enabled()); err != nil {
		return nil, err
	} else {
		this.cfg = cfg
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t *t) String() string {
	str := "<nginx"
	if version := t.Version(); version != "" {
		str += fmt.Sprintf(" version=%q", version)
	}
	if t.cmd != nil {
		str += fmt.Sprintf(" cmd=%q", t.cmd)
	}
	return str + ">"
}

func (s state) String() string {
	switch s {
	case stateNone:
		return "None"
	case stateRun:
		return "Run"
	case stateStop:
		return "Stop"
	case stateTerm:
		return "Term"
	case stateKill:
		return "Kill"
	case stateInfo:
		return "Info"
	case stateErr:
		return "Error"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (t *t) Run(ctx context.Context) error {
	var result error
	var wg sync.WaitGroup

	// Run the version command to get the nginx version string
	if version, err := NewWithCommand(t.cmd.Path(), "-v"); err != nil {
		return err
	} else {
		version.Err = func(_ *Cmd, data []byte) {
			t.version = append(t.version, data...)
		}
		if err := version.Run(); err != nil {
			return err
		}
	}

	// Add stdout, stderr for the nginx command
	t.cmd.Out = func(_ *Cmd, data []byte) {
		s := bytes.TrimSpace(data)
		t.Emit(event.Infof(ctx, stateInfo, string(s)))
	}
	t.cmd.Err = func(_ *Cmd, data []byte) {
		s := bytes.TrimSpace(data)
		t.Emit(event.Infof(ctx, stateInfo, string(s)))
	}
	t.test.Err = func(_ *Cmd, data []byte) {
		s := bytes.TrimSpace(data)
		t.Emit(event.Infof(ctx, stateInfo, string(s)))
	}

	// Create a cancelable context
	child, cancel := context.WithCancel(ctx)

	// Run nginx in the background, cancel when an error is returned
	wg.Add(1)
	go func() {
		defer wg.Done()
		t.Emit(event.Infof(ctx, stateRun, "Starting nginx"))
		if err := t.cmd.Run(); err != nil {
			result = multierror.Append(result, err)
			cancel()
		}
	}()

	// Wait for the context to be cancelled, indicate we're stopping the server
	<-child.Done()
	t.Emit(event.Infof(ctx, stateStop, "Stopping nginx"))

	// Start ticker for monitoring the process
	ticker := time.NewTimer(time.Second)
	state := stateStop
	defer ticker.Stop()
FOR_LOOP:
	for {
		select {
		case <-ticker.C:
			// Check that the process is still running
			pid := t.cmd.Pid()
			if t.cmd.Exited() {
				t.Emit(event.Infof(ctx, stateStop, "Process %d exited", pid))
				break FOR_LOOP
			} else {
				t.Emit(event.Infof(ctx, state, "Sending shutdown signal to process %d", pid))
			}
			// Send signal to process
			// https://docs.nginx.com/nginx/admin-guide/basic-functionality/runtime-control/
			switch state {
			case stateStop:
				t.Emit(event.Infof(ctx, stateStop, "Graceful shutdown"))
				if err := t.cmd.Signal(syscall.SIGQUIT); err != nil {
					result = multierror.Append(result, err)
				}
				state = stateTerm
			case stateTerm:
				t.Emit(event.Infof(ctx, stateStop, "Immediate shutdown"))
				if err := t.cmd.Signal(syscall.SIGTERM); err != nil {
					result = multierror.Append(result, err)
				}
				state = stateKill
			case stateKill:
				t.Emit(event.Infof(ctx, stateStop, "Kill shutdown"))
				if err := t.cmd.Signal(syscall.SIGKILL); err != nil {
					result = multierror.Append(result, err)
				}
			}
			// Reset the timer
			ticker.Reset(5 * time.Second)
		}
	}

	// Wait for nginx to exit
	wg.Wait()

	// Return any errors
	return result
}

// Test configuration
func (t *t) Test() error {
	return t.test.Run()
}

// Test the configuration and then reload it (the SIGHUP signal)
func (t *t) Reload() error {
	if err := t.test.Run(); err != nil {
		return err
	}
	return t.cmd.Signal(syscall.SIGHUP)
}

// Reopen log files (the SIGUSR1 signal)
func (t *t) Reopen() error {
	return t.cmd.Signal(syscall.SIGUSR1)
}

// Version returns the nginx version string
func (t *t) Version() string {
	if t.version == nil {
		return ""
	} else {
		return string(bytes.TrimSpace(t.version))
	}
}
