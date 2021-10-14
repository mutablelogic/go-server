package main

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	// Packages
	router "github.com/mutablelogic/go-server/pkg/httprouter"
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

type Handler struct {
	fs     fs.FS
	prefix string
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	indexPage = "/index.html"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewFileSystemHandler(fs fs.FS, prefix string) http.Handler {
	return &Handler{fs, prefix}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (f *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/") {
		r.URL.Path = "/" + r.URL.Path
	}
	name := r.URL.Path
	if f.prefix != "" {
		name = path.Join(f.prefix, name)
	}
	serveFile(w, r, f.fs, path.Clean(name), true)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Nicked this code from the net/http package
func serveFile(w http.ResponseWriter, r *http.Request, fs fs.FS, name string, shouldRedirect bool) {
	fmt.Println("SERVE", r.URL, name)
	// redirect .../index.html to .../
	if strings.HasSuffix(r.URL.Path, indexPage) {
		redirect(w, r, "./")
		return
	}

	// Open file
	f, err := fs.Open(name)
	if err != nil {
		router.ServeError(w, http.StatusNotFound, err.Error())
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		router.ServeError(w, http.StatusNotFound, err.Error())
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
		ff, err := fs.Open(index)
		if err == nil {
			defer ff.Close()
			dd, err := ff.Stat()
			if err == nil {
				name = index
				d = dd
				f = ff
			}
		}
	}

	// Still a directory, return unauthorized
	if d.IsDir() {
		router.ServeError(w, http.StatusUnauthorized, "Directory listing not allowed")
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
