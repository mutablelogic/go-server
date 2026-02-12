package types_test

import (
	"net/http"
	"testing"

	"github.com/mutablelogic/go-server/pkg/types"
	"github.com/stretchr/testify/assert"
)

func Test_RequestContentType(t *testing.T) {
	assert := assert.New(t)

	t.Run("JSON content type", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodPost, "/", nil)
		r.Header.Set("Content-Type", "application/json")
		ct, err := types.RequestContentType(r)
		assert.NoError(err)
		assert.Equal("application/json", ct)
	})

	t.Run("Content type with charset", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodPost, "/", nil)
		r.Header.Set("Content-Type", "text/html; charset=utf-8")
		ct, err := types.RequestContentType(r)
		assert.NoError(err)
		assert.Equal("text/html", ct)
	})

	t.Run("Empty content type", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		_, err := types.RequestContentType(r)
		assert.Error(err)
	})

	t.Run("Multipart form data", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodPost, "/", nil)
		r.Header.Set("Content-Type", "multipart/form-data; boundary=abc123")
		ct, err := types.RequestContentType(r)
		assert.NoError(err)
		assert.Equal("multipart/form-data", ct)
	})
}

func Test_AcceptContentType(t *testing.T) {
	assert := assert.New(t)

	t.Run("JSON accept", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Accept", "application/json")
		ct, err := types.AcceptContentType(r)
		assert.NoError(err)
		assert.Equal("application/json", ct)
	})

	t.Run("Wildcard accept", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Accept", "*/*")
		ct, err := types.AcceptContentType(r)
		assert.NoError(err)
		assert.Equal("*/*", ct)
	})

	t.Run("Empty accept", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		_, err := types.AcceptContentType(r)
		assert.Error(err)
	})
}
