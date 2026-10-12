package tui

import (
	"io"
	"os"
	"strings"

	// Packages
	glamour "github.com/charmbracelet/glamour"
	gansi "github.com/charmbracelet/glamour/ansi"
	gstyles "github.com/charmbracelet/glamour/styles"
	lipgloss "github.com/charmbracelet/lipgloss"
	term "github.com/charmbracelet/x/term"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Markdown is text which is rendered as markdown, for example as a table cell
type Markdown string

type markdown struct {
	width    int
	renderer *glamour.TermRenderer
}

const (
	markdownDefaultWidth = 80
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// MarkdownFormatter returns a formatter which renders markdown documents.
// The style is chosen based on whether stdout is a terminal, and its
// background color.
func MarkdownFormatter(options ...Opt) *markdown {
	var opts opts
	for _, opt := range options {
		if opt != nil {
			opt(&opts)
		}
	}

	width := markdownDefaultWidth
	if opts.width > 0 {
		width = opts.width
	}

	return &markdown{width: width}
}

// newMarkdownCell returns a renderer for markdown within a table cell, with
// no document margins and no word wrapping (the table wraps cells instead)
func newMarkdownCell() (*glamour.TermRenderer, error) {
	style := markdownStyle()
	margin := uint(0)
	style.Document.Margin = &margin
	style.Document.BlockPrefix = ""
	style.Document.BlockSuffix = ""
	return glamour.NewTermRenderer(
		glamour.WithStyles(style),
		glamour.WithWordWrap(0),
	)
}

///////////////////////////////////////////////////////////////////////////////
// RENDER

func (m *markdown) Write(w io.Writer, text string) (int, error) {
	if m.renderer == nil {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithStyles(markdownStyle()),
			glamour.WithWordWrap(m.width),
		)
		if err != nil {
			return 0, err
		}
		m.renderer = renderer
	}

	out, err := m.renderer.Render(text)
	if err != nil {
		return 0, err
	}
	return io.WriteString(w, out)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// markdownStyle returns a copy of the style config to use, which is
// plain when stdout is not a terminal
func markdownStyle() gansi.StyleConfig {
	switch {
	case !term.IsTerminal(os.Stdout.Fd()):
		return gstyles.NoTTYStyleConfig
	case lipgloss.HasDarkBackground():
		return gstyles.DarkStyleConfig
	default:
		return gstyles.LightStyleConfig
	}
}

// renderMarkdownCell renders markdown for a table cell, removing surrounding
// blank lines
func renderMarkdownCell(renderer *glamour.TermRenderer, text Markdown) (string, error) {
	out, err := renderer.Render(string(text))
	if err != nil {
		return "", err
	}
	return strings.Trim(out, "\n"), nil
}
