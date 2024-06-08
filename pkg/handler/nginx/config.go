package nginx

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	BinaryPath string            `hcl:"binary_path" description:"Path to nginx binary"`
	ConfigPath string            `hcl:"config" description:"Path to persistent configuration"`
	DataPath   string            `hcl:"data" description:"Path to ephermeral data directory"`
	LogPath    string            `hcl:"log" description:"Path to log directory"`
	LogRotate  time.Duration     `hcl:"log_rotate_period" description:"TODO: Period for log rotations (1d)"`
	LogKeep    time.Duration     `hcl:"log_keep_period" description:"TODO: Period for log deletions (28d)"`
	Env        map[string]string `hcl:"env" description:"Environment variables to set"`
	Directives map[string]string `hcl:"directives" description:"Directives to set in nginx configuration"`

	// Private fields
	deletePaths []string
}

// Check interfaces are satisfied
var _ server.Plugin = Config{}

////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName          = "nginx-handler"
	defaultExec          = "nginx"
	defaultConf          = "nginx.conf"
	defaultConfExt       = ".conf"
	defaultConfDirMode   = 0755
	defaultConfRecursive = true
	defaultLogPath       = "logs" // Relative to the ConfigPath
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Name returns the name of the service
func (Config) Name() string {
	return defaultName
}

// Description returns the description of the service
func (Config) Description() string {
	return "serves requests and provides responses over HTTP, HTTPS and FCGI"
}

// Create a new task from the configuration
func (c Config) New() (server.Task, error) {
	return New(&c)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the path to the nginx binary
func (c Config) ExecFile() string {
	if c.BinaryPath == "" {
		c.BinaryPath = defaultExec
	}
	return c.BinaryPath
}

// Return the path to the persistent data (configurations)
func (c *Config) ConfigDir() (string, error) {
	if c.ConfigPath == "" {
		if path, err := os.MkdirTemp("", "nginx-conf-"); err != nil {
			return "", err
		} else {
			c.ConfigPath = path
			c.deletePaths = append(c.deletePaths, path)
		}
	}
	if stat, err := os.Stat(c.ConfigPath); err != nil {
		return "", err
	} else if !stat.IsDir() {
		return "", ErrBadParameter.With("not a directory: ", c.ConfigPath)
	} else {
		return c.ConfigPath, nil
	}
}

// Return the path to temporary data (pid, sockets)
func (c *Config) DataDir() (string, error) {
	if c.DataPath == "" {
		if path, err := os.MkdirTemp("", "nginx-data-"); err != nil {
			return "", err
		} else {
			c.DataPath = path
			c.deletePaths = append(c.deletePaths, path)
		}
	}
	if stat, err := os.Stat(c.DataPath); err != nil {
		return "", err
	} else if !stat.IsDir() {
		return "", ErrBadParameter.With("not a directory: ", c.DataPath)
	} else {
		return c.DataPath, nil
	}
}

// Return the path to logging data, relative to the configuration directory
func (c Config) LogDir(configDir string) (string, error) {
	if c.LogPath == "" {
		c.LogPath = defaultLogPath
	}
	if !filepath.IsAbs(c.LogPath) {
		c.LogPath = filepath.Clean(filepath.Join(configDir, c.LogPath))
		if _, err := os.Stat(c.LogPath); os.IsNotExist(err) {
			if err := os.MkdirAll(c.LogPath, defaultConfDirMode); err != nil {
				return "", err
			}
		}
	}
	if stat, err := os.Stat(c.LogPath); err != nil {
		return "", err
	} else if !stat.IsDir() {
		return "", ErrBadParameter.With("not a directory: ", c.LogPath)
	} else {
		return c.LogPath, nil
	}
}

// Return the flags for the nginx binary
func (c Config) Flags(configDir, prefix string) []string {
	result := make([]string, 0, 5)

	// Switch off daemon setting, log initially to stderr
	result = append(result, "-e", "stderr", "-g", "daemon off;")

	// Add other directives
	for k, v := range c.Directives {
		result = append(result, "-g", fmt.Sprint(k, " ", string(v)+";"))
	}

	// Check for configuration file
	if configDir != "" {
		result = append(result, "-c", fmt.Sprintf("%s/%s", configDir, defaultConf))
	}

	// Check for prefix path
	if prefix != "" {
		result = append(result, "-p", prefix)
	}

	// Return flags
	return result
}
