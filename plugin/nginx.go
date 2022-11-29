package plugin

// Nginx provides a forked nginx instance, with the ability to reopen log files
// and reload the configuration
type Nginx interface {
	// Test configuration, return error if it fails
	Test() error

	// Test the configuration and then reload it (the SIGHUP signal)
	Reload() error

	// Reopen log files (the SIGUSR1 signal)
	Reopen() error
}
