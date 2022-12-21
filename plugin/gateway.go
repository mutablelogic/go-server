package plugin

import "context"

// Gateway provides a set of handlers for a router
type Gateway interface {
	// Register routes for the gateway with a router
	RegisterHandlers(context.Context, Router)

	// Return the middleware in order, which is called from left to right, then right to left,
	// on the serving of the route
	//Middleware() []string
}
