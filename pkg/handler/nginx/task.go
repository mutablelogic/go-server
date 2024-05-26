package nginx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	cmd "github.com/mutablelogic/go-server/pkg/handler/nginx/cmd"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type nginx struct {
	// The command to run nginx
	run *cmd.Cmd

	// The command to test nginx configuration
	test *cmd.Cmd

	// The configuration directory
	config string

	// The version string from nginx
	version []byte
}

// Check interfaces are satisfied
var _ server.Task = (*nginx)(nil)
var _ server.ServiceEndpoints = (*nginx)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new http server from the configuration
func New(c Config) (*nginx, error) {
	task := new(nginx)

	// Set configuration to a temporary directory
	if config, err := os.MkdirTemp("", "nginx-"); err != nil {
		return nil, err
	} else {
		task.config = config
	}

	// Create a new command to run the server
	if run, err := cmd.New(c.Path(), c.Flags(task.config, "")...); err != nil {
		return nil, err
	} else if test, err := cmd.New(c.Path(), c.Flags(task.config, "")...); err != nil {
		return nil, err
	} else {
		task.run = run
		task.test = test
		task.test.SetArgs("-t", "-q")
	}

	// Add the environment variables
	if err := task.run.SetEnv(c.Env); err != nil {
		return nil, err
	} else if err := task.test.SetEnv(c.Env); err != nil {
		return nil, err
	}

	// Return success
	return task, nil
}

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

	// Remove the temporary directory
	defer func() {
		if _, err := os.Stat(task.config); err == nil {
			if err := os.RemoveAll(task.config); err != nil {
				result = errors.Join(result, err)
			}
		}
	}()

	// We need to copy the configuration files to the temporary directory
	if err := fsCopyTo(task.config); err != nil {
		return err
	}

	// Get the version of nginx
	if version, err := task.getVersion(); err != nil {
		return err
	} else {
		task.version = version
	}

	// Add stdout, stderr for the nginx commands
	task.run.Out = func(data []byte) {
		s := bytes.TrimSpace(data)
		fmt.Println("exec stdout", string(s))
	}
	task.run.Err = func(data []byte) {
		s := bytes.TrimSpace(data)
		fmt.Println("exec stderr", string(s))
	}
	task.test.Out = func(data []byte) {
		s := bytes.TrimSpace(data)
		fmt.Println(string(s))
	}
	task.test.Err = func(data []byte) {
		s := bytes.TrimSpace(data)
		fmt.Println(string(s))
	}

	// Run the nginx server in the background
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := task.run.Run(); err != nil {
			result = errors.Join(result, err)
		}
	}()

	// Wait for the context to be cancelled
	<-ctx.Done()

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
			pid := task.run.Pid()
			if err := task.run.Signal(termSignal); err != nil {
				result = errors.Join(result, err)
			}
			switch termSignal {
			case syscall.SIGQUIT:
				fmt.Printf("Sending graceful shutdown signal to process %d\n", pid)
				termSignal = syscall.SIGTERM
				signalTicker.Reset(20 * time.Second) // Escalate after 20 seconds
			case syscall.SIGTERM:
				fmt.Printf("Sending immediate shutdown signal to process %d\n", pid)
				termSignal = syscall.SIGKILL
				signalTicker.Reset(5 * time.Second) // Escalate after 5 seconds
			case syscall.SIGKILL:
				fmt.Printf("Sending kill signal to process %d\n", pid)
				signalTicker.Reset(5 * time.Second) // Continue sending every 5 seconds
			}
		}
	}

	// Wait for nginx to exit
	wg.Wait()

	// Return any errors
	return result
}

// Test configuration
func (task *nginx) Test() error {
	return task.test.Run()
}

// Test the configuration and then reload it (the SIGHUP signal)
func (task *nginx) Reload() error {
	if err := task.test.Run(); err != nil {
		return err
	}
	return task.run.Signal(syscall.SIGHUP)
}

// Reopen log files (the SIGUSR1 signal)
func (task *nginx) Reopen() error {
	return task.run.Signal(syscall.SIGUSR1)
}

// Version returns the nginx version string
func (task *nginx) Version() string {
	if task.version == nil {
		if version, err := task.getVersion(); err == nil {
			task.version = version
			return string(bytes.TrimSpace(version))
		}
	}
	return string(bytes.TrimSpace(task.version))
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (task *nginx) getVersion() ([]byte, error) {
	var result []byte

	// Run the version command to get the nginx version string
	version, err := cmd.New(task.run.Path(), "-v")
	if err != nil {
		return nil, err
	}
	version.Err = func(data []byte) {
		result = append(result, data...)
	}
	if err := version.Run(); err != nil {
		return nil, err
	}

	// Return success
	return result, nil
}
