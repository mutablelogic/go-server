package main

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	// Modules
	router "github.com/djthorpe/go-server/pkg/httprouter"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *template) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Open file/directory for reading
	file, err := this.filefs.Open(filepath.Join(".", req.URL.Path))
	if os.IsNotExist(err) {
		router.ServeError(w, http.StatusNotFound)

		// Delete document from cache
		/*
			go func() {
				if err := this.indexer.Delete(req.URL.Path); err != nil {
					this.log.Printf(req.Context(), "Unable to delete from indexer: %q: %v", req.URL.Path, err)
				}
			}()
		*/

		return
	} else if err != nil {
		// Some other error occured
		router.ServeError(w, http.StatusInternalServerError)
		this.log.Printf(req.Context(), "ServeHTTP: %q: %v", req.URL.Path, err)
		return
	}
	defer file.Close()

	// Obtain file information
	stat, err := file.Stat()
	if err != nil {
		router.ServeError(w, http.StatusInternalServerError)
		this.log.Printf(req.Context(), "ServeHTTP: %q: %v", req.URL.Path, err)
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
		if strings.HasSuffix(req.URL.Path, "/") == false {
			req.URL.Path = req.URL.Path + "/"
			http.Redirect(w, req, req.URL.String(), http.StatusPermanentRedirect)
			return
		}
		this.ServeDir(w, req, file.(fs.ReadDirFile), stat)
	} else if stat.Mode().IsRegular() {
		this.ServeFile(w, req, file, stat)
		return
	}

	// If we reached here, something went wrong
	router.ServeError(w, http.StatusInternalServerError)
	this.log.Printf(req.Context(), "Unable to serve %q", req.URL.Path)
}

func (this *template) ServeFile(w http.ResponseWriter, req *http.Request, file fs.File, info fs.FileInfo) {
	// Forbidden if file is hidden file
	if strings.HasPrefix(info.Name(), ".") {
		router.ServeError(w, http.StatusForbidden)
		this.log.Printf(req.Context(), "Forbidden to serve %q", req.URL.Path)
		return
	}
	router.ServeText(w, "Serve File "+info.Name(), http.StatusOK)

	/*
		// Obtain renderer by file extension
		var renderer nginx.Renderer
		if ext := filepath.Ext(info.Name()); ext != "" {
			renderer = this.renderers.Lookup(ext)
		}

		// TODO Obtain renderer by file mimetype

		// Report error if no renderer available
		if renderer == nil {
			server.ServeError(w, http.StatusNotImplemented)
			this.log.Printf(req.Context(), "No renderer for %q", req.URL.Path)
			return
		}

		// Render document
		doc, err := renderer.Render(req.Context(), file.(io.Reader), info)
		if err != nil {
			server.ServeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Get template
		tmpl, err := this.cache.Lookup(doc.Template(), this.tmplfs)
		if err != nil {
			server.ServeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if err := tmpl.Execute(w, doc); err != nil {
			this.log.Printf(req.Context(), "Template error for %q: %v", req.URL.Path, err)
			return
		}

		// Index document in background
		go func() {
			if err := this.indexer.Add(req.URL.Path, doc); err != nil {
				this.log.Printf(req.Context(), "Unable to index %q: %v", req.URL.Path, err)
			}
		}()
	*/
}

func (this *template) ServeDir(w http.ResponseWriter, req *http.Request, file fs.ReadDirFile, info fs.FileInfo) {
	// Forbidden if name is hidden folder
	if strings.HasPrefix(info.Name(), ".") {
		router.ServeError(w, http.StatusForbidden)
		this.log.Printf(req.Context(), "Forbidden to serve %q", req.URL.Path)
		return
	}
	router.ServeText(w, "Serve Dir "+info.Name(), http.StatusOK)
}
