package main

import (
	"context"
	"io"
	"io/fs"
	"strings"

	// Package imports
	document "github.com/mutablelogic/go-server/pkg/document"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p *plugin) Mimetypes() []string {
	return []string{"/"}
}

func (p *plugin) Read(context.Context, io.Reader, fs.FileInfo, map[DocumentKey]interface{}) (Document, error) {
	return nil, ErrNotImplemented.With("Read")
}

func (p *plugin) ReadDir(ctx context.Context, dir fs.ReadDirFile, info fs.FileInfo, meta map[DocumentKey]interface{}) (Document, error) {
	// Check arguments
	if info != nil && !info.Mode().IsDir() {
		return nil, ErrBadParameter.With("ReadDir")
	}

	// Create a document
	name := info.Name()
	if name == "." {
		name = "/"
	}
	doc := document.New(name, meta, "dir")
	doc.SetMeta(DocumentKeyDescription, "Directory")

	// Read directory
	entries, err := dir.ReadDir(-1)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		} else if info, err := entry.Info(); err != nil {
			return nil, err
		} else {
			doc.AppendFileInfo(info, info.Name())
		}
	}

	// Return document
	return doc, nil
}
