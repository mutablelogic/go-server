package nginx

import (
	"bytes"
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	// Packages

	cmd "github.com/mutablelogic/go-server/pkg/handler/nginx/cmd"
	folders "github.com/mutablelogic/go-server/pkg/handler/nginx/folders"
	provider "github.com/mutablelogic/go-server/pkg/provider"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type nginx struct {
	// The command to run nginx
	run *cmd.Cmd

	// The command to test nginx configuration
	test *cmd.Cmd

	// The persistent configuration path
	configPath string

	// The ephemeral data path
	dataPath string

	// The log path
	logPath string

	// The version string from nginx
	version []byte

	// The available and enabled configuration folders
	folders *folders.Config

	// The temporarily created paths which should be removed on exit
	deletePaths []string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new http server from the configuration
func New(c *Config) (*nginx, error) {
	task := new(nginx)

	// Set permissions on config and data dirs
	configDir, err := c.ConfigDir()
	if err != nil {
		return nil, err
	} else if err := os.Chmod(configDir, defaultConfDirMode); err != nil {
		return nil, err
	} else {
		task.configPath = configDir
	}
	dataDir, err := c.DataDir()
	if err != nil {
		return nil, err
	} else if err := os.Chmod(dataDir, defaultConfDirMode); err != nil {
		return nil, err
	} else {
		task.dataPath = dataDir
	}
	logDir, err := c.LogDir(configDir)
	if err != nil {
		return nil, err
	} else if err := os.Chmod(logDir, defaultConfDirMode); err != nil {
		return nil, err
	} else {
		task.logPath = logDir
	}

	// Create an available folder in the persistent data directory
	availablePath := filepath.Join(configDir, "available")
	if err := os.MkdirAll(availablePath, defaultConfDirMode); err != nil {
		return nil, err
	}

	// Create an enabled folder in the persistent data directory
	enabledPath := filepath.Join(configDir, "enabled")
	if err := os.MkdirAll(enabledPath, defaultConfDirMode); err != nil {
		return nil, err
	}

	// Read the configuration folders
	if folders, err := folders.New(availablePath, enabledPath, defaultConfExt, defaultConfRecursive); err != nil {
		return nil, err
	} else {
		folders.DirMode = defaultConfDirMode
		task.folders = folders
	}

	// Create a new command to run the server. Use prefix to ensure that
	// the document root is contained within the temporary directory
	if run, err := cmd.New(c.ExecFile(), c.Flags(task.configPath, task.configPath)...); err != nil {
		return nil, err
	} else if test, err := cmd.New(c.ExecFile(), c.Flags(task.configPath, task.configPath)...); err != nil {
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

	// Set the working directory to the ephemeral data directory
	task.run.SetDir(task.dataPath)
	task.test.SetDir(task.dataPath)

	// Set the paths which should be deleted on exit
	task.deletePaths = c.deletePaths

	// Return success
	return task, nil
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// return the persistent config path
func (nginx *nginx) ConfigPath() string {
	return nginx.configPath
}

// return the ephermeral data path
func (nginx *nginx) DataPath() string {
	return nginx.dataPath
}

// return the log path
func (nginx *nginx) LogPath() string {
	return nginx.logPath
}

// Test configuration
func (nginx *nginx) Test() error {
	return nginx.test.Run()
}

// Test the configuration and then reload it (the SIGHUP signal)
func (nginx *nginx) Reload() error {
	// Reload the folders
	if err := nginx.folders.Reload(); err != nil {
		return err
	}

	// Test the configuration
	if err := nginx.test.Run(); err != nil {
		return err
	}

	// Signal the server to reload
	return nginx.run.Signal(syscall.SIGHUP)
}

// Reopen log files (the SIGUSR1 signal)
func (nginx *nginx) Reopen() error {
	return nginx.run.Signal(syscall.SIGUSR1)
}

// Version returns the nginx version string
func (nginx *nginx) Version() string {
	if nginx.version == nil {
		if version, err := nginx.getVersion(); err == nil {
			nginx.version = version
			return string(bytes.TrimSpace(version))
		}
	}
	return string(bytes.TrimSpace(nginx.version))
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (nginx *nginx) log(ctx context.Context, line string) {
	line = strings.TrimSpace(line)
	if logger := provider.Logger(ctx); logger != nil {
		logger.Print(ctx, line)
	} else {
		log.Println(line)
	}
}

func (nginx *nginx) getVersion() ([]byte, error) {
	var result []byte

	// Run the version command to get the nginx version string
	version, err := cmd.New(nginx.run.Path(), "-v")
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
