package main

import (
	"context"
	"io"
	"io/fs"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p *plugin) Mimetypes() []string {
	return []string{"/"}
}

func (p *plugin) Read(ctx context.Context, r io.Reader, info fs.FileInfo) (Document, error) {
	return nil, ErrNotImplemented.With("Read")
}

func (p *plugin) ReadDir(ctx context.Context, dir fs.ReadDirFile, info fs.FileInfo) (Document, error) {
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
	return NewDocument("/", info, entries), nil
}
