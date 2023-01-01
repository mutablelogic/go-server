package static

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"sync"

	// Package imports
	iface "github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/httpserver/util"
	task "github.com/mutablelogic/go-server/pkg/task"
	plugin "github.com/mutablelogic/go-server/plugin"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type static struct {
	task.Task
	sync.RWMutex

	label  string
	prefix string
	fs     fs.FS
}

var _ plugin.Gateway = (*static)(nil)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	indexPage = "/index.html"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new static serving task, with a specific filesystem and path for the
// static files. If the path is empty, then the root of the filesystem is used.
func NewWithPlugin(p iface.Plugin, filesys fs.FS, prefix string) (*static, error) {
	static := new(static)
	static.label = p.Label()

	// Create the handler
	if filesys == nil {
		return nil, ErrBadParameter.With("fs")
	} else {
		static.fs = filesys
	}

	// Get path in filesystem
	static.prefix = prefix

	// Return success
	return static, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (static *static) String() string {
	str := "<httpserver-static"
	if label := static.Label(); label != "" {
		str += fmt.Sprintf(" label=%q", label)
	}
	if prefix := static.prefix; prefix != "" {
		str += fmt.Sprintf(" prefix=%q", prefix)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Label returns the label of the router
func (static *static) Label() string {
	return static.label
}

// Description returns the label of the router
func (static *static) Description() string {
	return "Serves static files"
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Register routes for router
func (static *static) RegisterHandlers(parent context.Context, router plugin.Router) {
	// GET /
	//   Return list of prefixes and their handlers
	router.AddHandler(parent, nil, static.ServeHTTP)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (static *static) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path
	if static.prefix != "" {
		name = path.Join(static.prefix, name)
	}
	serveFile(w, r, static.fs, path.Clean(name), true)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Nicked this code from the net/http package
func serveFile(w http.ResponseWriter, r *http.Request, filesystem fs.FS, name string, shouldRedirect bool) {
	// redirect .../index.html to .../
	if strings.HasSuffix(r.URL.Path, indexPage) {
		redirect(w, r, "./")
		return
	}

	// Open file
	f, err := filesystem.Open(name)
	if errors.Is(err, fs.ErrNotExist) {
		util.ServeError(w, http.StatusNotFound, err.Error())
		return
	} else if errors.Is(err, fs.ErrPermission) {
		util.ServeError(w, http.StatusForbidden, err.Error())
		return
	} else if err != nil {
		util.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		util.ServeError(w, http.StatusNotFound, err.Error())
		return
	}

	if shouldRedirect {
		// redirect to canonical path: / at end of directory url
		// r.URL.Path always begins with /
		url := r.URL.Path
		if d.IsDir() {
			if url[len(url)-1] != '/' {
				redirect(w, r, path.Base(url)+"/")
				return
			}
		} else {
			if url[len(url)-1] == '/' {
				redirect(w, r, "../"+path.Base(url))
				return
			}
		}
	}

	if d.IsDir() {
		url := r.URL.Path
		// redirect if the directory name doesn't end in a slash
		if url == "" || url[len(url)-1] != '/' {
			redirect(w, r, path.Base(url)+"/")
			return
		}

		// use contents of index.html for directory, if present
		index := strings.TrimSuffix(name, "/") + indexPage
		ff, err := filesystem.Open(index)
		if err == nil {
			defer ff.Close()
			dd, err := ff.Stat()
			if err == nil {
				d = dd
				f = ff
			}
		}
	}

	// Still a directory, return unauthorized
	if d.IsDir() {
		util.ServeError(w, http.StatusUnauthorized, "Directory listing not allowed")
		return
	}

	// Serve content
	http.ServeContent(w, r, d.Name(), d.ModTime(), f.(io.ReadSeeker))
}

// Moved Permanently response
func redirect(w http.ResponseWriter, r *http.Request, path string) {
	if q := r.URL.RawQuery; q != "" {
		path += "?" + q
	}
	w.Header().Set("Location", path)
	w.WriteHeader(http.StatusMovedPermanently)
}
