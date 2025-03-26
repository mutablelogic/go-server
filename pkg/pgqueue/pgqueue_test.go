package pgqueue_test

import (
	"testing"

	// Packages
	test "github.com/djthorpe/go-pg/pkg/test"
)

// Global connection variable
var conn test.Conn

// Start up a container and test the pool
func TestMain(m *testing.M) {
	test.Main(m, &conn)
}
