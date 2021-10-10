package main

import (
	"bufio"
	"context"
	"io"
	"io/fs"
	"strings"

	// Package imports
	"github.com/zyedidia/highlight"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p *plugin) Mimetypes() []string {
	return []string{"text/plain", ".txt", ".md"}
}

func (p *plugin) Read(ctx context.Context, r io.Reader, info fs.FileInfo, meta map[DocumentKey]interface{}) (Document, error) {
	// Check arguments
	if info != nil && !info.Mode().IsRegular() {
		return nil, ErrBadParameter.With("Read")
	}

	// Create a document
	doc := NewDocument(info.Name())
	data := []string{}
	var def *highlight.Def

	// Read in lines, detect content type from first non-blank line
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if scanner.Text() == "" {
			doc.append(strings.Join(data, "\n"))
			data = data[:0]
			continue
		}
		if def == nil {
			def = highlight.DetectFiletype(p.defs, info.Name(), []byte(scanner.Text()))
		}
		data = append(data, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(data) > 0 {
		doc.append(strings.Join(data, "\n"))
	}

	// Highlight the document

	// Set def metadata
	if def != nil && def.FileType != "Unknown" {
		doc.addTag(def.FileType)
	}

	// Apply additional metadata
	for k, v := range meta {
		doc.setMeta(k, v)
	}

	// Return document
	return doc, nil
}

func (p *plugin) ReadDir(context.Context, fs.ReadDirFile, fs.FileInfo, map[DocumentKey]interface{}) (Document, error) {
	return nil, ErrNotImplemented.With("ReadDir")
}
