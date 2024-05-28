package nginx

import (
	// Packages
	"fmt"

	server "github.com/mutablelogic/go-server"
)

////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	BinaryPath string            `hcl:"binary_path" description:"Path to nginx binary"`
	Env        map[string]string `hcl:"env" description:"Environment variables to set"`
	Directives map[string]string `hcl:"directives" description:"Directives to set in nginx configuration"`
	Available  string            `hcl:"available" description:"Path to available configuration files"`
	Enabled    string            `hcl:"enabled" description:"Path to enabled configuration files"`
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
	if c.BinaryPath == "" {
		c.BinaryPath = defaultExec
	}
	return New(c)
}

// Return the path to the nginx binary
func (c Config) Path() string {
	if c.BinaryPath == "" {
		c.BinaryPath = defaultExec
	}
	return c.BinaryPath
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
