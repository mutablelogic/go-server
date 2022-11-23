package plugin

// Gateway provides a set of routes and middleware which is applied
// to those routes
type Gateway interface {
	// Return the prefix for this gateway
	Prefix() string

	// Return the middleware in order, which is called from left to right, then right to left,
	// on the serving of the route
	Middleware() []string
}
