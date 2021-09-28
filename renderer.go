package server

import (
	"context"
	"html/template"
	"io"
	"io/fs"
	"time"
)

/////////////////////////////////////////////////////////////////////
// TEMPLATE & INDEXER INTERFACES

// Renderer translates a data stream into a document
type Renderer interface {
	Plugin

	// Return default mimetypes and file extensions handled by this renderer
	Mimetypes() []string

	// Render a file into a document, with reader and optional file info
	Read(context.Context, io.Reader, fs.FileInfo) (Document, error)

	// Render a directory into a document, with optional file info
	ReadDir(context.Context, fs.ReadDirFile, fs.FileInfo) (Document, error)
}

// Document to be rendered or indexed
type Document interface {
	Title() string
	Description() string
	Shortform() template.HTML
	Tags() []string
	File() DocumentFile
	Meta() map[DocumentKey]interface{}
	HTML() []DocumentSection
}

// DocumentFile represents a document materialized on a file system
type DocumentFile interface {
	// Name of the file, including the file extension
	Name() string

	// Parent path for the file, not including the filename
	Path() string

	// Extension for the file, including the "."
	Ext() string

	// Last modification time for the document
	ModTime() time.Time

	// Size in bytes
	Size() int64
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
	DocumentKeyMimetype    DocumentKey = "mimetype"
)
