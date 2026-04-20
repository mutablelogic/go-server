package openapi

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func TestParseMarkdownSections(t *testing.T) {
	const input = `# Top

Preamble text.

## Section A

Body of A.

### Subsection A1

Body of A1.

## Section B

Body of B.
`

	doc := ParseMarkdown([]byte(input))

	t.Run("Sections", func(t *testing.T) {
		assert := assert.New(t)
		sections := doc.Sections()
		assert.Len(sections, 4)

		assert.Equal(1, sections[0].Level)
		assert.Equal("Top", sections[0].Title)
		assert.Contains(sections[0].Body, "Preamble text.")

		assert.Equal(2, sections[1].Level)
		assert.Equal("Section A", sections[1].Title)

		assert.Equal(3, sections[2].Level)
		assert.Equal("Subsection A1", sections[2].Title)

		assert.Equal(2, sections[3].Level)
		assert.Equal("Section B", sections[3].Title)
	})

	t.Run("SectionLookup", func(t *testing.T) {
		assert := assert.New(t)
		s := doc.Section(2, "Section A")
		assert.Contains(s.Body, "Body of A.")

		s = doc.Section(2, "section a")
		assert.Equal("Section A", s.Title)

		s = doc.Section(2, "Nonexistent")
		assert.Empty(s.Title)
	})
}

func TestParseMarkdownPreamble(t *testing.T) {
	const input = `Some text before any heading.

# First
Content.
`
	doc := ParseMarkdown([]byte(input))
	assert := assert.New(t)

	sections := doc.Sections()
	assert.Len(sections, 2)

	assert.Equal(0, sections[0].Level)
	assert.Equal("", sections[0].Title)
	assert.Contains(sections[0].Body, "Some text before any heading.")

	assert.Equal(1, sections[1].Level)
	assert.Equal("First", sections[1].Title)
}

func TestParseMarkdownClosingHashes(t *testing.T) {
	doc := ParseMarkdown([]byte("## Foo ##\n\nbar"))
	s := doc.Section(2, "Foo")
	assert.Contains(t, s.Body, "bar")
}

func TestParseMarkdownInlineCodeHeading(t *testing.T) {
	doc := ParseMarkdown([]byte("## `GET /auth/config`\n\nbody"))
	s := doc.Section(2, "GET /auth/config")
	assert.Equal(t, "GET /auth/config", s.Title)
	assert.Equal(t, "body", s.Body)
}

func TestParseMarkdownEmpty(t *testing.T) {
	doc := ParseMarkdown(nil)
	assert.Len(t, doc.Sections(), 1)
	assert.Equal(t, 0, doc.Sections()[0].Level)
}
