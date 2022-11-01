package httpserver

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	// Package imports
	"github.com/hashicorp/go-multierror"
	fcgi "github.com/mutablelogic/go-server/pkg/httpserver/fcgi"
	task "github.com/mutablelogic/go-server/pkg/task"
	plugin "github.com/mutablelogic/go-server/plugin"
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
func NewWithPlugin(p Plugin) (*t, error) {
	this := new(t)

	if isHostPort(p.listen) {
		// Create net server
		if http, err := netserver(p.listen, p.tls, p.timeout, p.router.(http.Handler)); err != nil {
			return nil, err
		} else {
			this.http = http
		}
	} else if abs, err := filepath.Abs(p.listen); err != nil {
		return nil, err
	} else {
		// Check listen for being (host, port). If not, then run as FCGI server
		if fcgi, err := fcgiserver(abs, p.router.(http.Handler)); err != nil {
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

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		if err := t.stop(); err != nil {
			result = multierror.Append(result, err)
		}
	}()
	if err := t.runInForeground(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		result = multierror.Append(result, err)
	}
	wg.Wait()
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// TODO: return true if:
//  :<number|name> <ip4>:<number> <ip6>:<number> <iface>:<number|name>
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

func fcgiserver(path string, handler http.Handler) (*fcgi.Server, error) {
	fcgi := new(fcgi.Server)
	fcgi.Network = "unix"
	fcgi.Addr = path
	fcgi.Handler = handler

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
