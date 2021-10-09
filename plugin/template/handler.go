package main

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	// Modules
	router "github.com/mutablelogic/go-server/pkg/httprouter"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *templates) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Open file/directory for reading
	file, err := this.filefs.Open(filepath.Join(".", req.URL.Path))
	if os.IsNotExist(err) {
		router.ServeError(w, http.StatusNotFound)
		return
	} else if err != nil {
		// Some other error occured
		router.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer file.Close()

	// Obtain file information
	stat, err := file.Stat()
	if err != nil {
		router.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Set headers
	modified := stat.ModTime()
	w.Header().Set("Last-Modified", modified.Format(http.TimeFormat))

	// Return not-modified
	if ifmodified := req.Header.Get("If-Modified-Since"); ifmodified != "" {
		if date, err := time.Parse(http.TimeFormat, ifmodified); err == nil {
			if modified.Before(date) {
				router.ServeError(w, http.StatusNotModified)
				return
			}
		}
	}

	// Serve as folder
	if stat.IsDir() {
		if !strings.HasSuffix(req.URL.Path, "/") {
			req.URL.Path = req.URL.Path + "/"
			http.Redirect(w, req, req.URL.String(), http.StatusPermanentRedirect)
			return
		}
		this.ServeDir(w, req, file.(fs.ReadDirFile), stat)
		return
	} else if stat.Mode().IsRegular() {
		this.ServeFile(w, req, file, stat)
		return
	}

	// If we reached here, something went wrong
	router.ServeError(w, http.StatusInternalServerError, "Unable to serve: ", req.URL.Path)
}

func (this *templates) ServeFile(w http.ResponseWriter, req *http.Request, file fs.File, info fs.FileInfo) {
	// Forbidden if file is hidden file
	if strings.HasPrefix(info.Name(), ".") {
		router.ServeError(w, http.StatusForbidden, "Forbidden to serve: ", req.URL.Path)
		return
	}

	// Render document
	doc, err := this.Read(req.Context(), file.(io.Reader), "", info)
	if err != nil {
		router.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get template
	tmpl, err := this.Cache.Lookup(this.Default)
	if err != nil {
		router.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Render document through the template
	if err := tmpl.Execute(w, doc); err != nil {
		// TODO: Print out error
		fmt.Println("TODO Template error for %q: %v", req.URL.Path, err)
	}
}

func (this *templates) ServeDir(w http.ResponseWriter, req *http.Request, file fs.ReadDirFile, info fs.FileInfo) {
	// Forbidden if name is hidden folder
	if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
		router.ServeError(w, http.StatusForbidden, "Forbidden to serve: ", req.URL.Path)
		return
	}

	// Render document
	doc, err := this.ReadDir(req.Context(), file, info)
	if err != nil {
		router.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get template
	tmpl, err := this.Cache.Lookup(this.Default)
	if err != nil {
		router.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Render document through the template
	if err := tmpl.Execute(w, doc); err != nil {
		// TODO: Print out error
		fmt.Println("TODO Template error for %q: %v", req.URL.Path, err)
	}
}
