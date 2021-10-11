package document

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

type Doc struct {
	name     string
	meta     map[DocumentKey]interface{}
	tags     []string
	sections []DocumentSection
}

type fileinfo struct {
	fs.FileInfo
	anchor string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(name string, meta map[DocumentKey]interface{}, tags ...string) *Doc {
	doc := new(Doc)

	// Set up document
	doc.name = name
	doc.meta = make(map[DocumentKey]interface{})

	// Add meta and tags
	for k, v := range meta {
		doc.SetMeta(k, v)
	}
	for _, tag := range tags {
		doc.AddTag(tag)
	}

	// Return success
	return doc
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (d *Doc) String() string {
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
	for k, v := range d.Meta() {
		switch v := v.(type) {
		case string:
			str += fmt.Sprintf(" %s=%q", k, v)
		default:
			str += fmt.Sprintf(" %s=%v", k, v)
		}
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (d *Doc) Title() string {
	if title, ok := d.meta[DocumentKeyTitle].(string); ok {
		return title
	} else {
		return d.name
	}
}

func (d *Doc) Description() string {
	if desc, ok := d.meta[DocumentKeyDescription].(string); ok {
		return desc
	} else {
		return ""
	}
}

func (d *Doc) Shortform() template.HTML {
	return ""
}

func (d *Doc) Tags() []string {
	return d.tags
}

func (d *Doc) Meta() map[DocumentKey]interface{} {
	return d.meta
}

func (d *Doc) HTML() []DocumentSection {
	return d.sections
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - FILE INFO

func (s *fileinfo) Title() string {
	return s.Name()
}

func (s *fileinfo) Level() uint {
	return 0
}

func (s *fileinfo) HTML() template.HTML {
	return template.HTML("")
}

func (s *fileinfo) Anchor() string {
	return s.anchor
}

func (s *fileinfo) Class() string {
	return "fileinfo"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (d *Doc) SetMeta(key DocumentKey, value interface{}) {
	d.meta[key] = value
}

func (d *Doc) AddTag(key string) {
	key = strings.ToLower(key)
	if !sliceContains(d.tags, key) {
		d.tags = append(d.tags, key)
	}
}
func (d *Doc) AppendFileInfo(info fs.FileInfo, anchor string) {
	d.sections = append(d.sections, &fileinfo{info, anchor})
}
