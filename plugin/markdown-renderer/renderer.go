package main

import (
	"context"
	"io"
	"io/fs"
	"io/ioutil"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-server"

	// Package imports
	parser "github.com/gomarkdown/markdown/parser"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p *plugin) Mimetypes() []string {
	return []string{"text/markdown", ".md"}
}

func (p *plugin) Read(ctx context.Context, r io.Reader, info fs.FileInfo) (Document, error) {
	// Check arguments
	if info != nil && !info.Mode().IsRegular() {
		return nil, ErrBadParameter.With("Read")
	}

	// Parse markdown document
	parser := parser.NewWithExtensions(parser.CommonExtensions | parser.Titleblock | parser.AutoHeadingIDs | parser.Attributes)
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Return document
	return NewDocument("/", info, parser.Parse(data)), nil
}

func (p *plugin) ReadDir(context.Context, fs.ReadDirFile, fs.FileInfo) (Document, error) {
	return nil, ErrNotImplemented.With("ReadDir")
}
