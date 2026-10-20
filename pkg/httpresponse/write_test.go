package httpresponse_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/stretchr/testify/assert"
)

func Test_Empty(t *testing.T) {
	assert := assert.New(t)

	t.Run("OK", func(t *testing.T) {
		rec := httptest.NewRecorder()
		err := httpresponse.Empty(rec, http.StatusOK)
		assert.NoError(err)
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal("0", rec.Header().Get("Content-Length"))
		assert.Equal("", rec.Body.String())
	})

	t.Run("NoContent", func(t *testing.T) {
		rec := httptest.NewRecorder()
		err := httpresponse.Empty(rec, http.StatusNoContent)
		assert.NoError(err)
		assert.Equal(http.StatusNoContent, rec.Code)
		assert.Equal("0", rec.Header().Get("Content-Length"))
	})
}

func Test_Write(t *testing.T) {
	assert := assert.New(t)

	t.Run("WithBody", func(t *testing.T) {
		rec := httptest.NewRecorder()
		err := httpresponse.Write(rec, http.StatusOK, "text/plain", func(w io.Writer) (int, error) {
			n, err := w.Write([]byte("hello world"))
			return n, err
		})
		assert.NoError(err)
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal("text/plain", rec.Header().Get("Content-Type"))
		assert.Equal("hello world", rec.Body.String())
	})

	t.Run("NilFunc", func(t *testing.T) {
		rec := httptest.NewRecorder()
		err := httpresponse.Write(rec, http.StatusOK, "text/plain", nil)
		assert.NoError(err)
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal("", rec.Body.String())
	})

	t.Run("CodeZeroDefaultsToOK", func(t *testing.T) {
		rec := httptest.NewRecorder()
		err := httpresponse.Write(rec, 0, "text/plain", nil)
		assert.NoError(err)
		assert.Equal(http.StatusOK, rec.Code)
	})

	t.Run("FuncReturnsError", func(t *testing.T) {
		rec := httptest.NewRecorder()
		err := httpresponse.Write(rec, http.StatusOK, "text/plain", func(w io.Writer) (int, error) {
			return 0, io.ErrUnexpectedEOF
		})
		assert.Error(err)
		assert.Equal(http.StatusInternalServerError, rec.Code)
	})

	t.Run("PresetContentType", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rec.Header().Set("Content-Type", "application/json")
		err := httpresponse.Write(rec, http.StatusOK, "text/plain", nil)
		assert.NoError(err)
		assert.Equal("application/json", rec.Header().Get("Content-Type"))
	})

	t.Run("ZeroBytesWritten", func(t *testing.T) {
		rec := httptest.NewRecorder()
		err := httpresponse.Write(rec, http.StatusOK, "text/plain", func(w io.Writer) (int, error) {
			return 0, nil
		})
		assert.NoError(err)
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal("", rec.Body.String())
	})
}

func Test_JSON_ZeroCode(t *testing.T) {
	assert := assert.New(t)

	rec := httptest.NewRecorder()
	err := httpresponse.JSON(rec, 0, 0, map[string]string{"key": "value"})
	assert.NoError(err)
	assert.Equal(http.StatusInternalServerError, rec.Code)
}
