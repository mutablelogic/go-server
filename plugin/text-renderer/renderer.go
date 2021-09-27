package main

import (
	"bufio"
	"context"
	"io"
	"io/fs"
	"strings"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p *plugin) Mimetypes() []string {
	return []string{"text/plain", ".txt", ".md"}
}

func (p *plugin) Read(ctx context.Context, r io.Reader, info fs.FileInfo) (Document, error) {
	// Check arguments
	if info != nil && !info.Mode().IsRegular() {
		return nil, ErrBadParameter.With("Read")
	}

	// Create a document
	doc := NewDocument("/", info)
	data := []string{}

	// Read in lines
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if scanner.Text() == "" {
			doc.append(strings.Join(data, "\n"))
			data = data[:0]
		} else {
			data = append(data, scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(data) > 0 {
		doc.append(strings.Join(data, "\n"))
	}

	// Return document
	return doc, nil
}

func (p *plugin) ReadDir(context.Context, fs.ReadDirFile, fs.FileInfo) (Document, error) {
	return nil, ErrNotImplemented.With("ReadDir")
}
