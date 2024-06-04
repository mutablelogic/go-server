package auth

import (
	"context"

	// Packages
	server "github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/provider"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type auth struct {
	jar        TokenJar
	tokenBytes int
}

// Check interfaces are satisfied
var _ server.Task = (*auth)(nil)
var _ server.ServiceEndpoints = (*auth)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new auth task from the configuration
func New(c Config) (*auth, error) {
	task := new(auth)

	// Set token jar
	if c.TokenJar == nil {
		return nil, ErrInternalAppError.With("missing 'tokenjar'")
	} else {
		task.jar = c.TokenJar
	}

	// Set token bytes
	if c.TokenBytes <= 0 {
		task.tokenBytes = defaultTokenBytes
	} else {
		task.tokenBytes = c.TokenBytes
	}

	// Return success
	return task, nil
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the label
func (task *auth) Label() string {
	// TODO
	return defaultName
}

// Run the task until the context is cancelled
func (task *auth) Run(ctx context.Context) error {
	var result error

	// Logger
	logger := provider.Logger(ctx)

	// If there are no tokens, then create a "root" token
	if tokens := task.jar.Tokens(); len(tokens) == 0 {
		token := NewToken(defaultRootNme, task.tokenBytes, 0, ScopeRoot)
		logger.Printf(ctx, "Creating root token %q for scope %q", token.Value, ScopeRoot)
		if err := task.jar.Create(token); err != nil {
			return err
		}
	}

	// Run the task until cancelled
	<-ctx.Done()

	// Return any errors
	return result
}
