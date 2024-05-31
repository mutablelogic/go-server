package nginx

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Nginx interface {
	// test the configuration and return an error if it fails
	Test() error

	// test the configuration and then reload it (the SIGHUP signal)
	Reload() error

	// reopen log files (the SIGUSR1 signal)
	Reopen() error

	// return the nginx version string
	Version() string

	// return the configuration path
	Config() string
}
