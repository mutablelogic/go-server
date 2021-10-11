package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"io/ioutil"
	"strings"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"

	// Package imports
	ast "github.com/gomarkdown/markdown/ast"
	html "github.com/gomarkdown/markdown/html"
	parser "github.com/gomarkdown/markdown/parser"
	document "github.com/mutablelogic/go-server/pkg/document"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type section struct {
	bytes.Buffer
	level                uint
	title, class, anchor string
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p *plugin) Mimetypes() []string {
	return []string{"text/markdown", ".md"}
}

func (p *plugin) Read(ctx context.Context, r io.Reader, info fs.FileInfo, meta map[DocumentKey]interface{}) (Document, error) {
	// Check arguments
	if info != nil && !info.Mode().IsRegular() {
		return nil, ErrBadParameter.With("Read")
	}

	// Create a document
	doc := document.New(info.Name(), meta, "markdown")
	doc.SetMeta(DocumentKeyDescription, "Markdown Document")
	doc.SetMeta(DocumentKeyContentType, "text/markdown")

	// Parse markdown document
	parser := parser.NewWithExtensions(parser.CommonExtensions | parser.Titleblock | parser.AutoHeadingIDs | parser.Attributes)
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	parseMarkdown(parser.Parse(data), doc)

	// Return document
	return doc, nil
}

func (p *plugin) ReadDir(context.Context, fs.ReadDirFile, fs.FileInfo, map[DocumentKey]interface{}) (Document, error) {
	return nil, ErrNotImplemented.With("ReadDir")
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func parseMarkdown(root ast.Node, doc *document.Doc) {
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
				appendSection(doc, section)
			}
			if entering {
				if v.IsTitleblock {
					skip = setMetaFromTitleBlock(doc, v)
				} else {
					title := titleFromHeading(v)
					skip = true
					heading = true
					section.title = title
					section.level = uint(v.Level)
					section.anchor = v.HeadingID
				}
			}
		case *ast.BlockQuote:
			// If before non-titleblock heading, then use as shortform
			if entering && !heading {
				skip = setMetaFromBlockQuote(doc, v)
			}
		}
		if skip {
			return ast.SkipChildren
		} else {
			return renderer.RenderNode(section, node, entering)
		}
	})
	appendSection(doc, section)
}

func appendSection(doc *document.Doc, section *section) {
	if section.Len() > 0 {
		doc.AppendHTML(section.title, section.level, template.HTML(section.String()), "#"+section.anchor, section.class)
	}
	section.Truncate(0)
}

/*
func (d *document) setMeta(k DocumentKey, v interface{}) {
	switch k {
	case DocumentKey("tags"):
		d.appendTags(strings.FieldsFunc(v.(string), func(c rune) bool { return c == ',' || c == ' ' }))
	default:
		d.meta[k] = v
	}
}
*/

func setMetaFromTitleBlock(doc *document.Doc, node *ast.Heading) bool {
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
				key := DocumentKey(strings.TrimSpace(strings.ToLower(kv[0])))
				fmt.Println(key, "=>", strings.TrimSpace(kv[1]))
			}
			continue
		}
		/*
			if _, exists := doc.Meta()[DocumentKeyTitle]; !exists {
				d.appendString(DocumentKeyTitle, line, " ")
			} else {
				d.appendString(DocumentKeyDescription, line, " ")
			}
		*/
	}

	return true
}

func titleFromHeading(node *ast.Heading) string {
	// Create heading title
	var title bytes.Buffer
	for _, child := range node.Children {
		if leaf := child.AsLeaf(); leaf != nil {
			title.Write(leaf.Literal)
		}
	}
	// TODO: Set title if level one and not set yet
	//if _, exists := d.meta[DocumentKeyTitle]; exists {
	//	return false
	//}

	// Return heading
	return title.String()
}

func setMetaFromBlockQuote(doc *document.Doc, node *ast.BlockQuote) bool {
	// Ignore if shortform has already been set
	//if _, exists := d.meta[DocumentKeyShortform]; exists {
	//	return false
	//}

	// Render shortform
	renderer := html.NewRenderer(html.RendererOptions{
		Flags: html.CommonFlags,
	})
	buf := new(bytes.Buffer)
	ast.WalkFunc(node, func(node ast.Node, entering bool) ast.WalkStatus {
		return renderer.RenderNode(buf, node, entering)
	})

	// Set shortform
	fmt.Println("shortform", "=>", template.HTML(buf.String()))

	//	d.meta[DocumentKeyShortform] = template.HTML(buf.String())

	// Set skip
	return true
}
