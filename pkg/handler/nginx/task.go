package nginx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"syscall"

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

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const ()

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
	//var wg sync.WaitGroup
	var result error

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
		fmt.Println("test stdout", string(s))
	}
	task.test.Err = func(data []byte) {
		s := bytes.TrimSpace(data)
		fmt.Println("test stderr", string(s))
	}

FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			// TODO: Stop the server
			break FOR_LOOP
		}
	}

	// Delete the configuration directory
	if err := os.RemoveAll(task.config); err != nil {
		result = errors.Join(result, err)
	}

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
