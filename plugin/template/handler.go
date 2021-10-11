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

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type TemplateEnv struct {
	Document Document
	File     fs.FileInfo
	Ext      string
	Path     string
	Parent   string
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *templates) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Modify request
	if !strings.HasPrefix(req.URL.Path, "/") {
		req.URL.Path = "/" + req.URL.Path
	}

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
	info, err := file.Stat()
	if err != nil {
		router.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Set headers
	modified := info.ModTime()
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

	// Forbidden if file is hidden file
	if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
		router.ServeError(w, http.StatusForbidden, "Forbidden to serve: ", req.URL.Path)
		return
	}

	// Serve as folder
	if info.IsDir() {
		if !strings.HasSuffix(req.URL.Path, "/") {
			req.URL.Path = req.URL.Path + "/"
			http.Redirect(w, req, req.URL.String(), http.StatusPermanentRedirect)
			return
		}
		this.ServeDir(w, req, file.(fs.ReadDirFile), info)
		return
	}

	// Not a regular file
	if !info.Mode().IsRegular() {
		router.ServeError(w, http.StatusInternalServerError, "Unable to serve: ", req.URL.Path)
		return
	}

	// Detect content type
	mimetype, charset, err := this.DetectContentType(file, info)
	if err != nil {
		router.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Close and re-open file
	if err := file.Close(); err != nil {
		router.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	file, err = this.filefs.Open(filepath.Join(".", req.URL.Path))
	if err != nil {
		router.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Render document - TODO: Add support for charset
	doc, err := this.Read(req.Context(), file.(io.Reader), info, map[DocumentKey]interface{}{
		DocumentKeyContentType: mimetype,
		DocumentKeyCharset:     charset,
	})
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
	if err := tmpl.Execute(w, env(doc, info, req.URL.Path)); err != nil {
		router.ServeError(w, http.StatusBadGateway, err.Error())
	}
}

func (this *templates) ServeDir(w http.ResponseWriter, req *http.Request, file fs.ReadDirFile, info fs.FileInfo) {
	// Render document
	doc, err := this.ReadDir(req.Context(), file, info, nil)
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
	if err := tmpl.Execute(w, env(doc, info, req.URL.Path)); err != nil {
		router.ServeError(w, http.StatusBadGateway, err.Error())
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func env(doc Document, info fs.FileInfo, path string) *TemplateEnv {
	parent := filepath.Dir(strings.TrimSuffix(path, "/"))
	if parent == "." {
		parent = ""
	} else if !strings.HasSuffix(parent, "/") {
		parent = parent + "/"
	}
	return &TemplateEnv{
		Document: doc,
		File:     info,
		Ext:      filepath.Ext(info.Name()),
		Path:     path,
		Parent:   parent,
	}
}

func (t *TemplateEnv) FormatTime(d time.Time) string {
	return d.Format(time.RFC3339)
}

func (t *TemplateEnv) FormatRelTime(d time.Time) string {
	diff := time.Now().Sub(d)
	if diff < 0 {
		return t.FormatTime(d)
	}
	return fmt.Sprintf("%d ago", diff)
}
