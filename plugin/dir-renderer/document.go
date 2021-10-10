package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"strings"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type document struct {
	name    string
	entries []DocumentSection
	meta    map[DocumentKey]interface{}
}

type direntry struct {
	fs.DirEntry
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewDocument(name string, entries []fs.DirEntry, meta map[DocumentKey]interface{}) *document {
	document := new(document)
	if entries == nil {
		return nil
	}

	// Set up document
	document.name = name
	document.meta = make(map[DocumentKey]interface{})

	// Add documents array
	document.entries = make([]DocumentSection, 0, len(entries))
	for _, entry := range entries {
		// Ignore hidden files and folders
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		document.entries = append(document.entries, &direntry{entry})
	}

	// Add metadata
	for key, value := range meta {
		document.meta[key] = value
	}

	// Return success
	return document
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (d *document) String() string {
	str := "<document"
	if title := d.Title(); title != "" {
		str += fmt.Sprintf(" title=%q", title)
	}
	if desc := d.Description(); desc != "" {
		str += fmt.Sprintf(" description=%q", desc)
	}
	if shortform := d.Shortform(); shortform != "" {
		str += fmt.Sprintf(" shortform=%q", shortform)
	}
	if tags := d.Tags(); len(tags) > 0 {
		str += fmt.Sprintf(" tags=%q", tags)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (d *document) Title() string {
	if d.name == "." {
		return "/"
	} else {
		return d.name
	}
}

func (d *document) Description() string {
	return "Directory"
}

func (d *document) Shortform() template.HTML {
	return ""
}

func (d *document) Tags() []string {
	return []string{"dir"}
}

func (d *document) Meta() map[DocumentKey]interface{} {
	return d.meta
}

func (d *document) HTML() []DocumentSection {
	return d.entries
}

func (s *direntry) Title() string {
	if s.DirEntry.IsDir() {
		return s.DirEntry.Name() + "/"
	} else {
		return s.DirEntry.Name()
	}
}

func (s *direntry) Level() uint {
	return 0
}

func (s *direntry) HTML() template.HTML {
	return template.HTML("")
}

func (s *direntry) Anchor() string {
	if s.DirEntry.IsDir() {
		return s.DirEntry.Name() + "/"
	} else {
		return s.DirEntry.Name()
	}
}

func (s *direntry) Class() string {
	if s.DirEntry.IsDir() {
		return "dir"
	} else {
		return ""
	}
}
