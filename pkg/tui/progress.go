package tui

import (
	"fmt"
	"io"
	"math"
	"strings"

	// Packages
	lipgloss "github.com/charmbracelet/lipgloss"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type progress struct {
	width   int
	label   lipgloss.Style
	filled  lipgloss.Style
	empty   lipgloss.Style
	percent lipgloss.Style
}

// labelWidth is the fixed character width of the label column.
// barOverhead is the total non-bar characters in a full progress line:
//
//	label(labelWidth) + " [" + bar + "] " + percent(6) = labelWidth + 10
const (
	labelWidth  = 40
	barOverhead = labelWidth + 10 // " [" + "] " + percent(6)
	barMinWidth = 10
	barDefault  = 20
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func Progress(options ...Opt) *progress {
	var opts opts
	for _, opt := range options {
		if opt != nil {
			opt(&opts)
		}
	}

	// When SetWidth is given a total line width, derive the bar width from it.
	// Otherwise fall back to the default bar width.
	barWidth := barDefault
	if opts.width > barOverhead+barMinWidth {
		barWidth = opts.width - barOverhead
	}

	renderer := lipgloss.NewRenderer(nil)
	self := &progress{
		width: barWidth,
	}
	self.label = renderer.NewStyle().Inline(true)
	self.filled = renderer.NewStyle().Foreground(lipgloss.Color("10"))
	self.empty = renderer.NewStyle().Foreground(lipgloss.Color("8"))
	self.percent = renderer.NewStyle().Width(6).Align(lipgloss.Right)

	return self
}

///////////////////////////////////////////////////////////////////////////////
// RENDER

func (p *progress) Write(w io.Writer, status string, percent float64) (int, error) {
	status = formatStatus(status)
	percent = clampPercent(percent)

	if percent <= 0 && status == "" {
		return io.WriteString(w, "")
	}

	filled := min(int(math.Round((percent/100.0)*float64(p.width))), p.width)
	bar := p.filled.Render(strings.Repeat("█", filled)) + p.empty.Render(strings.Repeat("░", p.width-filled))
	if status == "" {
		return io.WriteString(w, fmt.Sprintf("[%s] %s", bar, p.percent.Render(fmt.Sprintf("%5.1f%%", percent))))
	}
	return io.WriteString(w, fmt.Sprintf("%s [%s] %s", p.label.Render(status), bar, p.percent.Render(fmt.Sprintf("%5.1f%%", percent))))
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func clampPercent(percent float64) float64 {
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}

func formatStatus(status string) string {
	status = strings.NewReplacer("\r", " ", "\n", " ", "\t", " ").Replace(status)
	status = strings.TrimRight(status, " ")
	if status == "" {
		return ""
	}

	runes := []rune(status)
	if len(runes) > labelWidth {
		status = string(runes[:labelWidth-1]) + "…"
	}

	if pad := labelWidth - lipgloss.Width(status); pad > 0 {
		status += strings.Repeat(" ", pad)
	}

	return status
}
