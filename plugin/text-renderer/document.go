package main

import (
	"fmt"
	"html"
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	// Namespace imports
	. "github.com/djthorpe/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type document struct {
	path     string
	info     fs.FileInfo
	sections []DocumentSection
}

type textsection struct {
	data string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewDocument(path string, info fs.FileInfo) *document {
	document := new(document)

	// Check parameters
	if info == nil {
		return nil
	}

	// Set up document
	document.path = path
	document.info = info

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
	return d.info.Name()
}

func (d *document) Description() string {
	return d.info.Name()
}

func (d *document) Shortform() template.HTML {
	return ""
}

func (d *document) Tags() []string {
	return []string{"text"}
}

func (d *document) File() DocumentFile {
	return d
}

func (d *document) Meta() map[DocumentKey]interface{} {
	return nil
}

func (d *document) HTML() []DocumentSection {
	return d.sections
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
		return ".txt"
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
	return d.info.Size()
}

func (s *textsection) Title() string {
	return ""
}

func (s *textsection) Level() uint {
	return 0
}

func (s *textsection) HTML() template.HTML {
	return template.HTML(
		"<pre>" + html.EscapeString(s.data) + "</pre>",
	)
}

func (s *textsection) Anchor() string {
	return ""
}

func (s *textsection) Class() string {
	return "text"
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (d *document) append(data string) {
	d.sections = append(d.sections, &textsection{data})
}
