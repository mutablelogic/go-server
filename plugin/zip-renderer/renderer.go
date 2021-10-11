package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p *plugin) Mimetypes() []string {
	return []string{"application/zip", "application/gzip", ".zip", ".tar", ".gz", ".tgz"}
}

func (p *plugin) Read(ctx context.Context, r io.Reader, info fs.FileInfo, meta map[DocumentKey]interface{}) (Document, error) {
	// Check arguments
	if info != nil && !info.Mode().IsRegular() {
		return nil, ErrBadParameter.With("Read")
	}

	// Create a document
	doc := NewDocument(info.Name())

	// Perform decompression
	switch {
	case isZip(info):
		if err := readZip(ctx, r, info.Size(), doc); err != nil {
			return nil, err
		}
		doc.addTag("zip")
	case isTarCompressed(info):
		gz, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		if err := readTar(ctx, gz, doc); err != nil {
			return nil, err
		}
		doc.addTag("tar")
		doc.addTag("gzip")
	case isTar(info):
		if err := readTar(ctx, r, doc); err != nil {
			return nil, err
		}
		doc.addTag("tar")
	default:
		return nil, ErrBadParameter.With("Read: %s", info.Name())
	}

	// Apply additional metadata
	for k, v := range meta {
		doc.setMeta(k, v)
	}
	doc.addTag("dir")

	// Return document
	return doc, nil
}

func (p *plugin) ReadDir(context.Context, fs.ReadDirFile, fs.FileInfo, map[DocumentKey]interface{}) (Document, error) {
	return nil, ErrNotImplemented.With("ReadDir")
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func isZip(info fs.FileInfo) bool {
	ext := strings.ToLower(filepath.Ext(info.Name()))
	return ext == ".zip"
}

func isTar(info fs.FileInfo) bool {
	ext := strings.ToLower(filepath.Ext(info.Name()))
	return ext == ".tar"
}

func isTarCompressed(info fs.FileInfo) bool {
	name := strings.ToLower(info.Name())
	return strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".tgz")
}

func readZip(ctx context.Context, r io.Reader, size int64, doc *document) error {
	zip, err := zip.NewReader(NewReaderAt(r), size)
	if err != nil {
		return err
	}
	for _, file := range zip.File {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		fmt.Printf("TODO: %v\n", file)
	}

	// Return success
	return nil
}

func readTar(ctx context.Context, r io.Reader, doc *document) error {
	tar := tar.NewReader(r)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		header, err := tar.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		doc.append(header.Name)
	}

	// Return success
	return nil
}
