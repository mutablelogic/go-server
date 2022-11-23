package httpserver

import (
	"context"
	"errors"
	"net/http"

	// Namespace imports
	. "github.com/mutablelogic/terraform-provider-nginx"

	// Modules
	multierror "github.com/hashicorp/go-multierror"
)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Run until done
func (r *httpserver) Run(ctx context.Context) error {
	var result error
	go func() {
		<-ctx.Done()
		if err := r.stop(); err != nil {
			result = multierror.Append(result, err)
		}
	}()
	if err := r.runInForeground(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		result = multierror.Append(result, err)
	}
	return result
}

// Return label
func (r *httpserver) Label() string {
	return r.label
}

// Return event channel. No events are sent by the httpserver
func (*httpserver) C() <-chan Event {
	return nil
}
