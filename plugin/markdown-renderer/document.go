package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	// Namespace imports
	. "github.com/mutablelogic/go-server"

	// Package imports
	ast "github.com/gomarkdown/markdown/ast"
	html "github.com/gomarkdown/markdown/html"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type document struct {
	path     string
	info     fs.FileInfo
	root     ast.Node
	meta     map[DocumentKey]interface{}
	sections []DocumentSection
	tags     map[string]string
}

type section struct {
	bytes.Buffer
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewDocument(path string, info fs.FileInfo, root ast.Node) *document {
	doc := new(document)

	// Check parameters
	if info == nil {
		return nil
	}

	// Set up document
	doc.path = path
	doc.info = info
	doc.root = root
	doc.meta = make(map[DocumentKey]interface{})
	doc.tags = make(map[string]string)

	// Set up renderer
	renderer := html.NewRenderer(html.RendererOptions{
		Flags: html.CommonFlags,
	})

	// Walk tree
	section := new(section)
	heading := false
	ast.WalkFunc(root, func(node ast.Node, entering bool) ast.WalkStatus {
		var skip bool
		switch v := node.(type) {
		case *ast.Heading:
			if entering {
				if v.IsTitleblock {
					skip = doc.setFromTitleBlock(v)
				} else {
					skip = doc.setFromHeading(v)
					heading = true
				}
			}
		case *ast.BlockQuote:
			// If before non-titleblock heading, then use as shortform
			if entering && !heading {
				skip = doc.setFromBlockQuote(v)
			}
		}
		if skip {
			return ast.SkipChildren
		} else {
			return renderer.RenderNode(section, node, entering)
		}
	})

	// Append markdown tag
	doc.appendTags([]string{"markdown"})
	doc.setMeta(string(DocumentKeyMimetype), "text/markdown")

	// Return success
	return doc
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
	if v, ok := d.meta[DocumentKeyTitle].(string); ok {
		return v
	} else {
		return d.info.Name()
	}
}

func (d *document) Description() string {
	if v, ok := d.meta[DocumentKeyDescription].(string); ok {
		return v
	} else {
		return ""
	}
}

func (d *document) Shortform() template.HTML {
	if v, ok := d.meta[DocumentKeyShortform].(template.HTML); ok {
		return v
	} else {
		return ""
	}
}

func (d *document) Tags() []string {
	tags := make([]string, 0, len(d.tags))
	for _, tag := range d.tags {
		tags = append(tags, tag)
	}
	return tags
}

func (d *document) File() DocumentFile {
	return d
}

func (d *document) Meta() map[DocumentKey]interface{} {
	return d.meta
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

func (s *section) Title() string {
	return ""
}

func (s *section) Level() uint {
	return 0
}

func (s *section) HTML() template.HTML {
	return ""
}

func (s *section) Anchor() string {
	return ""
}

func (s *section) Class() string {
	return "text"
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (d *document) appendString(key DocumentKey, v, sep string) {
	if str, ok := d.meta[key].(string); ok {
		d.meta[key] = str + sep + v
	} else {
		d.meta[key] = v
	}
}

func (d *document) appendTags(tags []string) {
	for _, tag := range tags {
		key := strings.TrimSpace(strings.ToLower(tag))
		d.tags[key] = strings.TrimSpace(tag)
	}
}

func (d *document) setMeta(k, v string) {
	key := DocumentKey(strings.TrimSpace(strings.ToLower(k)))
	switch key {
	case DocumentKey("tags"):
		d.appendTags(strings.FieldsFunc(v, func(c rune) bool { return c == ',' || c == ' ' }))
	default:
		d.meta[key] = strings.TrimSpace(v)
	}
}

func (d *document) setFromTitleBlock(node *ast.Heading) bool {
	data := new(bytes.Buffer)
	ast.WalkFunc(node, func(node ast.Node, entering bool) ast.WalkStatus {
		if entering {
			if leaf := node.AsLeaf(); leaf != nil {
				data.Write(leaf.Literal)
			}
		}
		return ast.GoToNext
	})

	var meta bool
	for _, line := range strings.Split(data.String(), "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "%"))
		if line == "" {
			meta = true
			continue
		} else if meta {
			if kv := strings.SplitN(line, ":", 2); len(kv) == 2 {
				d.setMeta(kv[0], kv[1])
			}
			continue
		}
		if _, exists := d.meta[DocumentKeyTitle]; !exists {
			d.appendString(DocumentKeyTitle, line, " ")
		} else {
			d.appendString(DocumentKeyDescription, line, " ")
		}
	}

	return true
}

func (d *document) setFromHeading(node *ast.Heading) bool {
	// Ignore everything except Heading Level One
	if node.Level != 1 {
		return false
	}
	// Ignore if title has already been set
	if _, exists := d.meta[DocumentKeyTitle]; exists {
		return false
	}
	// There should be a single child
	for _, child := range node.Children {
		if leaf := child.AsLeaf(); leaf != nil {
			d.appendString(DocumentKeyTitle, string(leaf.Literal), " ")
		}
	}
	// Skip children
	return true
}

func (d *document) setFromBlockQuote(node *ast.BlockQuote) bool {
	// Ignore if shortform has already been set
	if _, exists := d.meta[DocumentKeyShortform]; exists {
		return false
	}

	// Render shortform
	renderer := html.NewRenderer(html.RendererOptions{
		Flags: html.CommonFlags,
	})
	buf := new(bytes.Buffer)
	ast.WalkFunc(node, func(node ast.Node, entering bool) ast.WalkStatus {
		return renderer.RenderNode(buf, node, entering)
	})

	// Set shortform
	d.meta[DocumentKeyShortform] = template.HTML(buf.String())

	// Set skip
	return true
}
