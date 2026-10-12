package tui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkdownWrite(t *testing.T) {
	var buffer bytes.Buffer

	n, err := MarkdownFormatter().Write(&buffer, "# Title\n\n- alpha\n- beta")
	require.NoError(t, err)
	assert.Equal(t, len(buffer.String()), n)
	assert.Contains(t, buffer.String(), "Title")
	assert.Contains(t, buffer.String(), "• alpha")
	assert.Contains(t, buffer.String(), "• beta")
}

func TestMarkdownWriteWraps(t *testing.T) {
	var buffer bytes.Buffer

	_, err := MarkdownFormatter(SetWidth(20)).Write(&buffer, strings.Repeat("word ", 20))
	require.NoError(t, err)
	for _, line := range strings.Split(buffer.String(), "\n") {
		assert.LessOrEqual(t, len(strings.TrimRight(line, " ")), 20)
	}
}
