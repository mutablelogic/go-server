package main

import (
	"fmt"
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

type filesection struct {
	path string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewDocument(name string) *document {
	document := new(document)

	// Set up document
	document.name = name
	document.tags = make([]string, 0, 2)
	document.meta = make(map[DocumentKey]interface{})

	// Add tags
	document.addTag("archive")

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

func (s *filesection) Title() string {
	return s.path
}

func (s *filesection) Level() uint {
	return 0
}

func (s *filesection) HTML() template.HTML {
	return template.HTML("")
}

func (s *filesection) Anchor() string {
	return ""
}

func (s *filesection) Class() string {
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
func (d *document) append(path string) {
	d.sections = append(d.sections, &filesection{path})
}
