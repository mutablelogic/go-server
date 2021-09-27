package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	// Namespace imports
	. "github.com/djthorpe/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type document struct {
	path    string
	info    fs.FileInfo
	entries []DocumentSection
}

type direntry struct {
	fs.DirEntry
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewDocument(path string, info fs.FileInfo, entries []fs.DirEntry) *document {
	document := new(document)

	// Return nil if info is nil
	if info == nil || entries == nil {
		return nil
	}

	// Set up document
	document.path = path
	document.info = info

	// Add documents array
	document.entries = make([]DocumentSection, 0, len(entries))
	for _, entry := range entries {
		// Ignore hidden files and folders
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		document.entries = append(document.entries, &direntry{entry})
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
	if name := d.Name(); name != "" {
		str += fmt.Sprintf(" name=%q", name)
	}
	if path := d.Path(); path != "" {
		str += fmt.Sprintf(" path=%q", path)
	}
	if ext := d.Ext(); ext != "" {
		str += fmt.Sprintf(" ext=%q", ext)
	}
	if modtime := d.ModTime(); !modtime.IsZero() {
		str += fmt.Sprint(" modtime=", modtime.Format(time.RFC3339))
	}
	if size := d.Size(); size > 0 {
		str += fmt.Sprint(" size=", size)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (d *document) Title() string {
	if d.info.Name() == "." {
		return "/"
	} else {
		return d.info.Name()
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

func (d *document) File() DocumentFile {
	if d.info != nil {
		return d
	} else {
		return nil
	}
}

func (d *document) Meta() map[DocumentKey]interface{} {
	return nil
}

func (d *document) HTML() []DocumentSection {
	return d.entries
}

func (d *document) Name() string {
	return d.Title()
}

func (d *document) Path() string {
	return d.path
}

func (d *document) Ext() string {
	if d.info != nil {
		return filepath.Ext(d.info.Name())
	} else {
		return "/"
	}
}

func (d *document) ModTime() time.Time {
	if d.info != nil {
		return d.info.ModTime()
	} else {
		return time.Time{}
	}
}

func (d *document) Size() int64 {
	return 0
}

func (s *direntry) Title() string {
	return s.DirEntry.Name()
}

func (s *direntry) Level() uint {
	return 0
}

func (s *direntry) HTML() template.HTML {
	return template.HTML("")
}

func (s *direntry) Anchor() string {
	return ""
}

func (s *direntry) Class() string {
	if s.DirEntry.IsDir() {
		return "dir"
	} else {
		return ""
	}
}
