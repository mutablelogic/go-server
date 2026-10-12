package tui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testRow struct {
	value string
}

func (r testRow) Header() []string {
	return []string{"Value"}
}

func (r testRow) Cell(i int) CellText {
	return r.value
}

func (r testRow) Width(i int) int {
	return 0
}

func TestTableWriteEmpty(t *testing.T) {
	var buffer bytes.Buffer

	n, err := TableFor[testRow]().Write(&buffer)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, "", buffer.String())
}

func TestTableWriteEndsWithNewline(t *testing.T) {
	var buffer bytes.Buffer

	table := TableFor[testRow]()
	table.Append(testRow{value: "alpha"})
	n, err := table.Write(&buffer)
	require.NoError(t, err)
	assert.Equal(t, len(buffer.String()), n)
	assert.True(t, strings.HasSuffix(buffer.String(), "\n"))
	assert.Contains(t, buffer.String(), "alpha")
}

func TestTableWriteFlushesBufferedRows(t *testing.T) {
	var buffer bytes.Buffer

	table := TableFor[testRow]()
	table.Append(testRow{value: "alpha"}, testRow{value: "beta"})

	_, err := table.Write(&buffer)
	require.NoError(t, err)
	assert.Contains(t, buffer.String(), "alpha")
	assert.Contains(t, buffer.String(), "beta")

	buffer.Reset()
	_, err = table.Write(&buffer)
	require.NoError(t, err)
	assert.Equal(t, "", buffer.String())
}

type testMarkdownRow struct {
	value Markdown
	count int
}

func (r testMarkdownRow) Header() []string {
	return []string{"Value", "Count"}
}

func (r testMarkdownRow) Cell(i int) CellText {
	if i == 0 {
		return r.value
	}
	return r.count
}

func (r testMarkdownRow) Width(i int) int {
	return 0
}

func TestTableWriteMarkdownCell(t *testing.T) {
	var buffer bytes.Buffer

	_, err := TableFor[testMarkdownRow]().Write(&buffer, testMarkdownRow{value: "- alpha\n- beta", count: 42})
	require.NoError(t, err)
	assert.Contains(t, buffer.String(), "• alpha")
	assert.Contains(t, buffer.String(), "• beta")
	assert.Contains(t, buffer.String(), "42")
	t.Log("\n" + buffer.String())
}
