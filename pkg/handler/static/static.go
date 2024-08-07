package static

import (
	"context"
	"errors"
	"html"
	"io"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	FS         fs.FS  `hcl:"fs" description:"File system to serve"`
	DirPrefix  string `hcl:"prefix" description:"Directory to serve files from"`
	DirListing bool   `hcl:"dir" description:"Serve directory listings"`
	Path       string `hcl:"path" description:"host/path to serve files on"`
}

type static struct {
	fs     fs.FS
	prefix string
	path   string
	dir    bool
}

// Ensure interfaces is implemented
var _ http.Handler = (*static)(nil)
var _ server.Plugin = Config{}
var _ server.Task = (*static)(nil)
var _ server.ServiceEndpoints = (*static)(nil)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName = "static"
	indexPage   = "/index.html"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Name returns the name of the service
func (Config) Name() string {
	return defaultName
}

// Description returns the description of the service
func (Config) Description() string {
	return "handler to serve static files"
}

// Create a new static handler from the configuration
func (c Config) New() (server.Task, error) {
	s := new(static)

	// Check file system
	if c.FS == nil {
		return nil, ErrBadParameter.Withf("fs is nil")
	} else {
		s.fs = c.FS
	}

	// Check directory prefix
	if c.DirPrefix != "" {
		c.DirPrefix = path.Clean(c.DirPrefix)
		if f, err := s.fs.Open(c.DirPrefix); err != nil {
			return nil, err
		} else if info, err := f.Stat(); err != nil {
			return nil, err
		} else if !info.IsDir() {
			return nil, ErrBadParameter.With("not a directory prefix: ", c.DirPrefix)
		} else {
			s.prefix = c.DirPrefix
		}
	}

	// Set other options
	s.dir = c.DirListing
	s.path = c.Path

	// Return success
	return s, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Implement the http.Handler interface to serve files
func (static *static) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serveFile(w, r, static.fs, static.prefix, path.Clean(r.URL.Path), static.dir, true)
}

// Run the static handler until the context is cancelled
func (static *static) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Return the label for the task
func (static *static) Label() string {
	// TODO
	return defaultName
}

// Add endpoints to the router
func (static *static) AddEndpoints(ctx context.Context, router server.Router) {
	router.AddHandler(ctx, static.path, static, http.MethodGet)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Serve a file from the file system
func serveFile(w http.ResponseWriter, r *http.Request, filesystem fs.FS, prefix, name string, dir bool, shouldRedirect bool) {
	// redirect .../index.html to .../
	if strings.HasSuffix(r.URL.Path, indexPage) {
		redirect(w, r, "./")
		return
	}

	// Add prefix and then fudge
	name = filepath.Join(prefix, name)
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}
	if name == "" {
		name = "."
	}

	// Open file
	f, err := filesystem.Open(name)
	if errors.Is(err, fs.ErrNotExist) {
		httpresponse.Error(w, http.StatusNotFound, err.Error())
		return
	} else if errors.Is(err, fs.ErrPermission) {
		httpresponse.Error(w, http.StatusForbidden, err.Error())
		return
	} else if err != nil {
		httpresponse.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		httpresponse.Error(w, http.StatusNotFound, err.Error())
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
		if dir {
			if err := serveDir(w, filesystem, name); err != nil {
				httpresponse.Error(w, http.StatusInternalServerError, err.Error())
			}
			return
		} else {
			httpresponse.Error(w, http.StatusUnauthorized, "Directory listing not allowed")
			return
		}
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

// serveDir lists the directory
func serveDir(w http.ResponseWriter, filesystem fs.FS, path string) error {
	entries, err := fs.ReadDir(filesystem, path)
	if err != nil {
		return err
	}

	// Modify the path
	if path == "." {
		path = "/"
	}

	// Write header
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<html><head><title>Index of " + path + "</title></head><body><h1>Index of " + path + "</h1><ul>"))

	// Add parent directory
	if path != "/" {
		w.Write([]byte("<li><a href=\"../\">../</a></li>"))
	}

	// Add child directories, ignoring any hidden files and directories
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if entry.IsDir() {
			name += "/"
		}
		w.Write([]byte("<li><a href=\"" + html.EscapeString(name) + "\">" + html.EscapeString(name) + "</a></li>"))
	}

	// Write footer
	w.Write([]byte("</ul></body></html>"))

	// Return success
	return nil
}
