package tui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTableSummaryNoResults(t *testing.T) {
	assert.Equal(t, "No scopes returned", TableSummary("scopes", 0, 0, nil).String())
}

func TestTableSummaryAllResults(t *testing.T) {
	limit := uint64(10)
	assert.Equal(t, "Showing all 3 scopes", TableSummary("scopes", 3, 0, &limit).String())
}

func TestTableSummaryPageRange(t *testing.T) {
	limit := uint64(25)
	assert.Equal(t, "Showing scopes 26-50 of 100", TableSummary("scopes", 100, 25, &limit).String())
}

func TestTableSummaryWriteEndsWithNewline(t *testing.T) {
	var buffer bytes.Buffer
	limit := uint64(10)

	n, err := TableSummary("scopes", 3, 0, &limit).Write(&buffer)
	require.NoError(t, err)
	assert.Equal(t, len(buffer.String()), n)
	assert.True(t, strings.HasSuffix(buffer.String(), "\n"))
	assert.Equal(t, "Showing all 3 scopes\n", buffer.String())
}
