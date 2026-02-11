package httpstatic_test

import (
	"testing"
	"testing/fstest"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpstatic "github.com/mutablelogic/go-server/pkg/httpstatic"
	"github.com/stretchr/testify/assert"
)

func Test_New_001(t *testing.T) {
	assert := assert.New(t)
	h, err := httpstatic.New("/static", fstest.MapFS{}, "Files", "Serve static files")
	assert.NoError(err)
	assert.NotNil(h)
	var _ server.HTTPFileServer = h
}

func Test_HandlerPath_001(t *testing.T) {
	assert := assert.New(t)
	h, err := httpstatic.New("/static", fstest.MapFS{}, "", "")
	assert.NoError(err)
	assert.Equal("/static", h.HandlerPath())
}

func Test_HandlerPath_002(t *testing.T) {
	assert := assert.New(t)
	h, err := httpstatic.New("assets/", fstest.MapFS{}, "", "")
	assert.NoError(err)
	assert.Equal("/assets", h.HandlerPath())
}

func Test_HandlerPath_003(t *testing.T) {
	assert := assert.New(t)
	h, err := httpstatic.New("", fstest.MapFS{}, "", "")
	assert.NoError(err)
	assert.Equal("/", h.HandlerPath())
}

func Test_HandlerFS_001(t *testing.T) {
	assert := assert.New(t)
	fs := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<h1>Hello</h1>")},
	}
	h, err := httpstatic.New("/", fs, "", "")
	assert.NoError(err)
	assert.NotNil(h.HandlerFS())
	data, err := h.HandlerFS().Open("index.html")
	assert.NoError(err)
	assert.NotNil(data)
	data.Close()
}

func Test_HandlerSpec_001(t *testing.T) {
	assert := assert.New(t)
	h, err := httpstatic.New("/docs", fstest.MapFS{}, "Documentation", "Serve documentation files")
	assert.NoError(err)
	spec := h.Spec()
	assert.NotNil(spec)
	assert.Equal("Documentation", spec.Summary)
	assert.Equal("Serve documentation files", spec.Description)
	assert.NotNil(spec.Get)
	assert.Equal("Documentation", spec.Get.Summary)
	assert.Equal("Serve documentation files", spec.Get.Description)
}

func Test_HandlerSpec_002(t *testing.T) {
	assert := assert.New(t)
	h, err := httpstatic.New("/", fstest.MapFS{}, "", "")
	assert.NoError(err)
	spec := h.Spec()
	assert.NotNil(spec)
	assert.Empty(spec.Summary)
	assert.Empty(spec.Description)
	assert.NotNil(spec.Get)
	assert.Empty(spec.Get.Summary)
	assert.Empty(spec.Get.Description)
}
