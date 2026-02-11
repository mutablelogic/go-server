package httpstatic

import (
	"io/fs"

	// Packages
	server "github.com/mutablelogic/go-server"
	openapi "github.com/mutablelogic/go-server/pkg/openapi/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type httpstatic struct {
	path        string
	fs          fs.FS
	summary     string
	description string
}

var _ server.HTTPFileServer = (*httpstatic)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(path string, fs fs.FS, summary, description string) (*httpstatic, error) {
	return &httpstatic{
		path:        types.NormalisePath(path),
		fs:          fs,
		summary:     summary,
		description: description,
	}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (h *httpstatic) HandlerPath() string {
	return h.path
}

func (h *httpstatic) HandlerFS() fs.FS {
	return h.fs
}

func (h *httpstatic) Spec() *openapi.PathItem {
	return &openapi.PathItem{
		Summary:     types.Ptr(h.summary),
		Description: types.Ptr(h.description),
		Get: &openapi.Operation{
			Summary:     types.Ptr(h.summary),
			Description: types.Ptr(h.description),
		},
	}
}
