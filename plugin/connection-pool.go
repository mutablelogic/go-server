package plugin

import (
	"github.com/mutablelogic/go-accessory"
)

// ConnectionPool provides a mechanism to get and put database connections
// from a pool
type ConnectionPool interface {
	// Get returns a connection from the pool
	Get() accessory.Conn

	// Put returns a connection to the pool
	Put(accessory.Conn)

	// Size returns the size of the pool
	Size() int
}
