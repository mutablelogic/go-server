package ldap

import (
	"context"
	"errors"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	provider "github.com/mutablelogic/go-server/pkg/provider"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Check interfaces are satisfied
var _ server.Task = (*ldap)(nil)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the label
func (task *ldap) Label() string {
	// TODO
	return defaultName
}

// Run the task until the context is cancelled
func (task *ldap) Run(ctx context.Context) error {
	var result error
	var first bool

	// Getting logging object
	log := provider.Logger(ctx)

	// Connect
	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()

FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		case <-timer.C:
			// Connect or ping connection
			if err := task.Connect(); err != nil {
				log.Print(ctx, err)

				// If this is the first time and we have bad credentials, then return the error
				if errors.Is(err, ErrNotAuthorized) && !first {
					result = err
					break FOR_LOOP
				}
			} else {
				// Indicate that we are connected. If subsequent connections fail, then
				// we will log the error but continue to try to connect
				if !first {
					log.Printf(ctx, "Connected to %s:%d", task.Host(), task.Port())
					first = true
				}
			}

			// We attempt to ping the connection every minute
			timer.Reset(deltaPingTime)
		}
	}

	// Return any errors
	return result
}

/*

func (self *ldap) Run(ctx context.Context) error {
	var result error

	// Bind to the connection, and re-bind occasionally if it fails
	state := stNone
	delta := durationMinBackoff
	timer := time.NewTimer(delta)

FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			break FOR_LOOP
		case <-timer.C:
			var err error
			state, delta, err = self.changeState(state, delta)
			if err != nil {
				// TODO: Emit error
				fmt.Fprintln(os.Stderr, err)
			}
			timer.Reset(delta)
		}
	}

	// Disconnect from LDAP connection
	if self.conn != nil {
		if err := ldapDisconnect(self.conn); err != nil {
			result = errors.Join(result, err)
		}
		self.conn = nil
	}

	// Return any errors
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (self *ldap) changeState(state connState, delta time.Duration) (connState, time.Duration, error) {
	var result error

	// Lock operation
	self.Mutex.Lock()
	defer self.Mutex.Unlock()

	// Connect, disconnect, bind and ping
	switch state {
	case stNone:
		// Disconnect
		if self.conn != nil {
			if err := ldapDisconnect(self.conn); err != nil {
				result = errors.Join(result, err)
			}
		}
		// Connect
		if conn, err := ldapConnect(self.Host(), self.Port(), self.tls); err != nil {
			delta = min(delta*2, durationMaxBackoff)
			result = errors.Join(result, err)
		} else {
			self.conn = conn
			state = stConnected
			delta = durationMinBackoff
		}
	case stConnected:
		if self.conn == nil {
			result = errors.Join(result, ErrInternalAppError.With("state is connected but connection is nil"))
			state = stNone
			delta = durationMinBackoff
		} else if err := ldapBind(self.conn, self.dn, self.password); err != nil {
			result = errors.Join(result, err)
			delta = min(delta*2, durationMaxBackoff)
		} else {
			state = stAuthenticated
			delta = durationMinBackoff
		}
	case stAuthenticated:
		// TODO: Check connection is still valid
		delta = min(delta*2, durationMaxBackoff)
	default:
		result = errors.Join(result, ErrInternalAppError.With("changeState"))
	}

	// Return any errors
	return state, delta, result
}

*/
