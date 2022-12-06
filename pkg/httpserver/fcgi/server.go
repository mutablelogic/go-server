package fcgi

// This file implements FastCGI from the perspective of a child process.

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
)

// A Server defines parameters for running an FCGI server.
// The zero value for Server is a valid configuration, which responds
// to requests over stdin
type Server struct {
	// Network and Addr is the address or path to the socket, which
	// are used to create the listener
	Network, Addr string

	// handler to invoke, http.DefaultServeMux if nil
	Handler http.Handler

	// Owner, Group and Mode are used to set the permissions on the socket
	// or -1 if no owner or group should be set
	Owner, Group int

	// File mode for the socket or 0 if no work should be done
	Mode os.FileMode

	// Private variables to flag shutdown
	listener net.Listener
	ctx      context.Context
	cancel   context.CancelFunc
}

func (s *Server) ListenAndServe() error {
	var wg sync.WaitGroup

	// Remove existing socket
	if (s.Network == "unix" || s.Network == "") && s.Addr != "" {
		// Check for existing file and remove it. Cannot use a directory
		// as a socket
		if stat, err := os.Stat(s.Addr); os.IsNotExist(err) {
			// File does not exist, so no nothing
		} else if err != nil {
			return err
		} else if stat.IsDir() {
			return fmt.Errorf("Cannot use an existing directory")
		} else if err := os.Remove(s.Addr); err != nil {
			return err
		}
	}

	// If Network and Addr are empty, use os.Stdin
	if (s.Network == "unix" || s.Network == "") && s.Addr == "" {
		if l, err := net.FileListener(os.Stdin); err != nil {
			return err
		} else {
			s.listener = l
		}
	} else {
		if l, err := net.Listen(s.Network, s.Addr); err != nil {
			return err
		} else {
			s.listener = l
		}
	}
	defer s.listener.Close()

	// Set default handler
	if s.Handler == nil {
		s.Handler = http.DefaultServeMux
	}

	// Set owner, group and mode
	if s.Owner >= 0 || s.Group >= 0 {
		if err := os.Chown(s.Addr, s.Owner, s.Group); err != nil {
			return err
		}
	}

	// Set mode
	if s.Mode > 0 {
		if err := os.Chmod(s.Addr, s.Mode); err != nil {
			return err
		}
	}

	// Set up semapore which when closed ends the loop
	s.ctx, s.cancel = context.WithCancel(context.Background())

	// Continue accepting requests until shutdown
FOR_LOOP:
	for {
		select {
		case <-s.ctx.Done():
			break FOR_LOOP
		default:
			rw, err := s.listener.Accept()
			if neterr, ok := err.(*net.OpError); ok && neterr.Err == net.ErrClosed {
				break FOR_LOOP
			} else if err != nil {
				return err
			}
			c := newChild(rw, s.Handler)
			wg.Add(1)
			go func() {
				c.serve()
				wg.Done()
			}()
		}
	}

	// Wait until all connections served
	wg.Wait()

	// Return success
	return nil
}

func (s *Server) Close() error {
	var result error
	if s.cancel != nil {
		result = s.listener.Close()
		s.cancel()
	}
	return result
}
