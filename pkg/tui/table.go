package tui

import (
	"fmt"
	"io"

	// Packages
	lipgloss "github.com/charmbracelet/lipgloss"
	lgtable "github.com/charmbracelet/lipgloss/table"
)

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

type TableRow interface {
	Header() []string
	Cell(int) string
	Width(int) int
}

///////////////////////////////////////////////////////////////////////////////
// TYPES

type table[T TableRow] struct {
	table  *lgtable.Table
	widths []int
	rows   []T
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func TableFor[T TableRow](options ...Opt) *table[T] {
	var opts opts
	for _, opt := range options {
		if opt != nil {
			opt(&opts)
		}
	}

	renderer := lipgloss.NewRenderer(nil)
	self := new(table[T])
	renderedTable := lgtable.New().
		Border(lipgloss.NormalBorder()).
		Wrap(true).
		StyleFunc(func(row, col int) lipgloss.Style {
			style := renderer.NewStyle()
			if col >= 0 && col < len(self.widths) && self.widths[col] > 0 {
				style = style.MaxWidth(self.widths[col])
			}
			if row == lgtable.HeaderRow {
				return style.Bold(true)
			}
			return style.Padding(0, 1)
		})
	if opts.width > 0 {
		renderedTable.Width(opts.width)
	}

	self.table = renderedTable
	return self
}

///////////////////////////////////////////////////////////////////////////////
// RENDER

func (t *table[T]) Append(rows ...T) *table[T] {
	if len(rows) > 0 {
		t.rows = append(t.rows, rows...)
	}
	return t
}

func (t *table[T]) Write(w io.Writer, rows ...T) (int, error) {
	if len(rows) > 0 {
		t.Append(rows...)
	}

	t.table.ClearRows()
	if len(t.rows) == 0 {
		return io.WriteString(w, "")
	}

	headers := t.rows[0].Header()
	t.widths = make([]int, len(headers))
	for i := range headers {
		t.widths[i] = t.rows[0].Width(i)
	}
	t.table.Headers(headers...)
	for _, row := range t.rows {
		cells := make([]string, len(headers))
		for i := range headers {
			cells[i] = row.Cell(i)
		}
		t.table.Row(cells...)
	}
	defer func() {
		t.rows = nil
	}()

	return io.WriteString(w, fmt.Sprintln(t.table.Render()))
}
