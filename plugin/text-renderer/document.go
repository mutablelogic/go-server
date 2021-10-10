package main

import (
	"fmt"
	"html"
	"html/template"
	"strings"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type document struct {
	name     string
	sections []DocumentSection
	tags     []string
	meta     map[DocumentKey]interface{}
}

type textsection struct {
	data string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewDocument(name string) *document {
	document := new(document)

	// Set up document
	document.name = name
	document.tags = make([]string, 0, 2)
	document.meta = make(map[DocumentKey]interface{})

	// Add "TEXT" tag
	document.addTag("text")

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
	return d.name
}

func (d *document) Description() string {
	return d.name
}

func (d *document) Shortform() template.HTML {
	return ""
}

func (d *document) Tags() []string {
	return d.tags
}

func (d *document) Meta() map[DocumentKey]interface{} {
	return d.meta
}

func (d *document) HTML() []DocumentSection {
	return d.sections
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

func (d *document) setMeta(key DocumentKey, value interface{}) {
	d.meta[key] = value
}

func (d *document) addTag(key string) {
	key = strings.ToLower(key)
	if !sliceContains(d.tags, key) {
		d.tags = append(d.tags, key)
	}
}
func (d *document) append(data string) {
	d.sections = append(d.sections, &textsection{data})
}
