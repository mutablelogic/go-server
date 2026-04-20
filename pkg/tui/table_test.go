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

func (r testRow) Cell(i int) string {
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

	n, err := TableFor[testRow]().Write(&buffer, testRow{value: "alpha"})
	require.NoError(t, err)
	assert.Equal(t, len(buffer.String()), n)
	assert.True(t, strings.HasSuffix(buffer.String(), "\n"))
	assert.Contains(t, buffer.String(), "alpha")
}
