package httpserver

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	// Package imports
	multierror "github.com/hashicorp/go-multierror"
	fcgi "github.com/mutablelogic/go-server/pkg/httpserver/fcgi"
	task "github.com/mutablelogic/go-server/pkg/task"
	plugin "github.com/mutablelogic/go-server/plugin"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type t struct {
	task.Task
	plugin.Router
	fcgi *fcgi.Server
	http *http.Server
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func NewWithPlugin(p Plugin, router plugin.Router) (*t, error) {
	this := new(t)

	handler, ok := router.(http.Handler)
	if !ok {
		return nil, ErrInternalAppError.Withf("Router %v does not implement http.Handler", router)
	}

	listen := p.Listen()
	if isHostPort(listen) {
		// net socket parameters
		timeout := p.Timeout()
		tls, err := p.TLS()
		if err != nil {
			return nil, err
		}
		// Create net server
		if http, err := netserver(listen, tls, timeout, handler); err != nil {
			return nil, err
		} else {
			this.http = http
		}
	} else if abs, err := filepath.Abs(listen); err != nil {
		return nil, err
	} else {
		// file socket parameters
		if owner, err := p.Owner(); err != nil {
			return nil, err
		} else if group, err := p.Group(); err != nil {
			return nil, err
		} else if fcgi, err := fcgiserver(abs, owner, group, p.Mode(), handler); err != nil {
			return nil, err
		} else {
			this.fcgi = fcgi
		}
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t *t) String() string {
	str := "<httpserver"
	if t.fcgi != nil {
		str += fmt.Sprintf(" fcgi=%q", t.fcgi.Addr)
	} else {
		str += fmt.Sprintf(" addr=%q", t.http.Addr)
		if t.http.TLSConfig != nil {
			str += " tls=true"
		}
		if t.http.ReadHeaderTimeout != 0 {
			str += fmt.Sprintf(" read_timeout=%v", t.http.ReadHeaderTimeout)
		}
	}
	if t.Router != nil {
		str += fmt.Sprintf(" router=%v", t.Router)
	}
	return str + ">"
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (t *t) Run(ctx context.Context) error {
	var result error
	var wg sync.WaitGroup

	// Create cancelable context
	child, cancel := context.WithCancel(ctx)
	defer cancel()

	// Stop server on context cancel
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-child.Done()
		if err := t.stop(); err != nil {
			result = multierror.Append(result, err)
		}
	}()

	// Run server in foreground, cancel when done
	if err := t.runInForeground(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		result = multierror.Append(result, err)
		cancel()
	}

	// Wait for gorutines to finish
	wg.Wait()

	// Return any errors
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// TODO: return true if:
//
//	:<number|name> <ip4>:<number> <ip6>:<number> <iface>:<number|name>
func isHostPort(listen string) bool {
	// Empty string not acceptable
	if listen == "" {
		return false
	}
	// Check for host:port and if no error, return true
	if _, _, err := net.SplitHostPort(listen); err == nil {
		return true
	} else {
		return false
	}
}

func fcgiserver(path string, uid, gid int, mode os.FileMode, handler http.Handler) (*fcgi.Server, error) {
	fcgi := new(fcgi.Server)
	fcgi.Network = "unix"
	fcgi.Addr = path
	fcgi.Handler = handler
	fcgi.Owner = uid
	fcgi.Group = gid
	fcgi.Mode = mode

	// Return success
	return fcgi, nil
}

func netserver(addr string, config *tls.Config, timeout time.Duration, handler http.Handler) (*http.Server, error) {
	srv := new(http.Server)
	srv.Addr = addr
	srv.Handler = handler
	srv.ReadHeaderTimeout = timeout
	srv.IdleTimeout = timeout
	if config != nil {
		srv.TLSConfig = config
	}

	// Return success
	return srv, nil
}

func (t *t) runInForeground() error {
	if t.fcgi != nil {
		return t.fcgi.ListenAndServe()
	} else if t.http.TLSConfig != nil {
		return t.http.ListenAndServeTLS("", "")
	} else {
		return t.http.ListenAndServe()
	}
}

func (t *t) stop() error {
	if t.fcgi != nil {
		return t.fcgi.Close()
	} else {
		return t.http.Close()
	}
}
