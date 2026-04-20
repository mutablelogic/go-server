package openapi

import (
	"bytes"
	_ "embed"
	"errors"
	"strings"

	// Packages
	tokenizer "github.com/mutablelogic/go-tokenizer"
	ast "github.com/mutablelogic/go-tokenizer/pkg/ast"
	markdown "github.com/mutablelogic/go-tokenizer/pkg/markdown"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type MarkdownDoc struct {
	root     ast.Node
	sections []Section
}

type Section struct {
	Level int
	Title string
	Body  string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func ParseMarkdown(data []byte) (doc *MarkdownDoc) {
	return &MarkdownDoc{
		root:     markdown.Parse(bytes.NewReader(data), tokenizer.Pos{}),
		sections: parseSections(string(data)),
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (doc *MarkdownDoc) Sections() []Section {
	if doc == nil {
		return nil
	}
	return doc.sections
}

func (doc *MarkdownDoc) Section(level int, title string) Section {
	if doc == nil {
		return Section{}
	}
	title = normalizeTitle(title)
	for _, section := range doc.sections {
		if section.Level == level && strings.EqualFold(section.Title, title) {
			return section
		}
	}
	return Section{}
}

// FindSection returns the first heading whose text matches title.
func (doc *MarkdownDoc) FindSection(title string) *markdown.Heading {
	if doc == nil || doc.root == nil {
		return nil
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return nil
	}

	var section *markdown.Heading
	_ = ast.Walk(doc.root, func(node ast.Node, depth int) error {
		heading, ok := node.(*markdown.Heading)
		if !ok {
			return nil
		}
		if strings.EqualFold(markdownNodeText(heading), title) {
			section = heading
			return errors.New("found markdown section")
		}
		return nil
	})
	return section
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func markdownNodeText(node ast.Node) string {
	if node == nil {
		return ""
	}

	var text strings.Builder
	_ = ast.Walk(node, func(node ast.Node, depth int) error {
		switch node := node.(type) {
		case *markdown.Text:
			text.WriteString(node.Value())
		case *markdown.Code:
			text.WriteString(node.String())
		}
		return nil
	})

	return strings.TrimSpace(text.String())
}

func parseSections(content string) []Section {
	lines := strings.Split(content, "\n")
	sections := make([]Section, 0, len(lines))

	var current *Section
	for _, line := range lines {
		if level, title, ok := parseHeading(line); ok {
			if current != nil {
				current.Body = strings.Trim(current.Body, "\n")
				sections = append(sections, *current)
			}
			current = &Section{Level: level, Title: title}
			continue
		}
		if current == nil {
			current = &Section{}
		}
		current.Body += line + "\n"
	}
	if current != nil {
		current.Body = strings.Trim(current.Body, "\n")
		sections = append(sections, *current)
	}

	return sections
}

func parseHeading(line string) (level int, title string, ok bool) {
	trimmed := strings.TrimLeft(line, " ")
	if len(trimmed) == 0 || trimmed[0] != '#' {
		return 0, "", false
	}
	i := 0
	for i < len(trimmed) && trimmed[i] == '#' {
		i++
	}
	if i > 6 {
		return 0, "", false
	}
	if i < len(trimmed) && trimmed[i] != ' ' && trimmed[i] != '\t' {
		return 0, "", false
	}
	title = strings.TrimSpace(trimmed[i:])
	title = strings.TrimRight(title, "# ")
	title = normalizeTitle(title)
	return i, title, true
}

func normalizeTitle(title string) string {
	title = strings.TrimSpace(title)
	for _, marker := range []string{"`", "**", "__", "*", "_"} {
		if strings.HasPrefix(title, marker) && strings.HasSuffix(title, marker) && len(title) >= 2*len(marker) {
			title = strings.TrimSpace(title[len(marker) : len(title)-len(marker)])
		}
	}
	return title
}
