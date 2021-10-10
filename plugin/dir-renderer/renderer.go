package main

import (
	"context"
	"io"
	"io/fs"

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

	// Read directory
	entries, err := dir.ReadDir(-1)
	if err != nil {
		return nil, err
	}

	// Return document
	return NewDocument("/", info, entries, meta), nil
}
