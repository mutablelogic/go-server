package main

import (
	"bufio"
	"context"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"strings"

	// Package imports
	document "github.com/mutablelogic/go-server/pkg/document"
	highlight "github.com/zyedidia/highlight"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	escaperune = map[rune]string{
		'&':  "&amp;",
		'\'': "&#39;",
		'<':  "&lt;",
		'>':  "&gt;",
		'"':  "&#34;",
	}
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
	doc := document.New(info.Name(), meta, "text")
	doc.AddTag("text")

	// Create a document
	var def *highlight.Def
	var data string

	// Read in lines, detect content type from first non-blank line
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if def == nil {
			def = highlight.DetectFiletype(p.defs, info.Name(), []byte(scanner.Text()))
		}
		data = data + scanner.Text() + "\n"
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Return error if can't detect content type
	if def == nil {
		return nil, ErrUnexpectedResponse.With(info.Name, " unable to detect content type")
	} else if def.FileType != "Unknown" {
		doc.AddTag(def.FileType)
	}

	// Highlight text
	var result, span template.HTML
	var class string

	matches := highlight.NewHighlighter(def).HighlightString(data)
	for n, l := range strings.Split(data, "\n") {
		for m, r := range l {
			// Check if the group changed at the current position
			if group, ok := matches[n][m]; ok {
				class, span = set(class, group)
				if span != "" {
					result = result + span
				}
			}
			// Append rune to document
			if str, exists := escaperune[r]; exists {
				result = result + template.HTML(str)
			} else {
				result = result + template.HTML(r)
			}
		}
		// This is at a newline, but highlighting might have been turned off at the very end of the line so we should check that.
		if group, ok := matches[n][len(l)]; ok {
			class, span = set(class, group)
			if span != "" {
				result = result + span
			}
		}
		result += template.HTML("\n")
	}
	_, span = set(class, highlight.Groups[""])
	if span != "" {
		result = result + span
	}

	// Append text to document
	doc.AppendHTML("", 0, template.HTML("<pre>"+result+"</pre>"), "", "text")

	// Return document
	return doc, nil
}

func (p *plugin) ReadDir(context.Context, fs.ReadDirFile, fs.FileInfo, map[DocumentKey]interface{}) (Document, error) {
	return nil, ErrNotImplemented.With("ReadDir")
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func set(class string, group highlight.Group) (string, template.HTML) {
	var result template.HTML

	n := fmt.Sprint(group)
	if class != n {
		if class != "" {
			result += template.HTML("</span>")
		}
		if n != "" {
			result += template.HTML(fmt.Sprintf("<span class=%q>", strings.Replace(n, ".", " ", -1)))
		}
	}
	return n, result
}
