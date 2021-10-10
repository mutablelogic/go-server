package server

import (
	"context"
	"html/template"
	"io"
	"io/fs"
)

/////////////////////////////////////////////////////////////////////
// TEMPLATE & INDEXER INTERFACES

// Renderer translates a data stream into a document
type Renderer interface {
	Plugin

	// Return default mimetypes and file extensions handled by this renderer
	Mimetypes() []string

	// Render a file into a document, with reader and optional file info and metadata
	Read(context.Context, io.Reader, fs.FileInfo, map[DocumentKey]interface{}) (Document, error)

	// Render a directory into a document, with optional file info and metadata
	ReadDir(context.Context, fs.ReadDirFile, fs.FileInfo, map[DocumentKey]interface{}) (Document, error)
}

// Document to be rendered or indexed
type Document interface {
	Title() string
	Description() string
	Shortform() template.HTML
	Tags() []string
	Meta() map[DocumentKey]interface{}
	HTML() []DocumentSection
}

// DocumentSection represents document HTML, split into sections
type DocumentSection interface {
	// Title for the section
	Title() string

	// Level of the section, or zero if no level defined
	Level() uint

	// Valid HTML for the section (which allows extraction into text tokens)
	HTML() template.HTML

	// Anchor name for the section, if any
	Anchor() string

	// Class name for the section, if any
	Class() string
}

// DocumentKey provides additional metadata for a document
type DocumentKey string

const (
	DocumentKeyTitle       DocumentKey = "title"
	DocumentKeyDescription DocumentKey = "description"
	DocumentKeyShortform   DocumentKey = "shortform"
	DocumentKeyAuthor      DocumentKey = "author"
	DocumentKeyArtwork     DocumentKey = "artwork"
	DocumentKeyThumbnail   DocumentKey = "thumbnail"
	DocumentKeyContentType DocumentKey = "mimetype"
	DocumentKeyCharset     DocumentKey = "charset"
)
