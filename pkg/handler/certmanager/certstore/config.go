package certstore

import (
	// Packages
	"os"
	"os/user"
	"strconv"

	server "github.com/mutablelogic/go-server"
)

////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	DataPath string `hcl:"data" description:"Data directory for certificates"`
	Group    string `hcl:"group" description:"Group ID for data directory"`
}

// Check interfaces are satisfied
var _ server.Plugin = Config{}

////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName     = "certstore"
	defaultDirMode  = os.FileMode(0700)
	defaultFileMode = os.FileMode(0600)
	groupDirMode    = os.FileMode(0050)
	groupFileMode   = os.FileMode(0060)
	allDirMode      = os.FileMode(0777)
	allFileMode     = os.FileMode(0666)
	certExt         = ".pem"
)

const (
	R_OK = 4
	W_OK = 2
	X_OK = 1
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Name returns the name of the service
func (Config) Name() string {
	return defaultName
}

// Description returns the description of the service
func (Config) Description() string {
	return "file-based storage for certificates and certificate authorities"
}

// Create a new task from the configuration
func (c Config) New() (server.Task, error) {
	return New(c)
}

// Return the gid of the group
func (c Config) GroupId() (int, error) {
	// No group change
	if c.Group == "" {
		return -1, nil
	}

	// Check for numerical group ID
	if gid, err := strconv.ParseUint(c.Group, 0, 32); err == nil {
		return int(gid), nil
	}

	// Check for group name
	if group, err := user.LookupGroup(c.Group); err != nil {
		return -1, err
	} else if gid, err := strconv.ParseUint(group.Gid, 0, 32); err != nil {
		return -1, err
	} else {
		return int(gid), nil
	}
}

// Return the dir mode for the data directory
func (c Config) DirMode() os.FileMode {
	dirMode := defaultDirMode
	if gid, err := c.GroupId(); gid != -1 && err == nil {
		dirMode |= groupDirMode
	}
	return (dirMode & allDirMode)
}

// Return the file mode for the data directory
func (c Config) FileMode() os.FileMode {
	dirMode := defaultFileMode
	if gid, err := c.GroupId(); gid != -1 && err == nil {
		dirMode |= groupFileMode
	}
	return (dirMode & allFileMode)
}
