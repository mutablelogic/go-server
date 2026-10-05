package httpresponse

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"
)

func Test_Error_001(t *testing.T) {
	assert := assert.New(t)

	notfound := Err(http.StatusNotFound)
	assert.NotNil(notfound)
	assert.Equal(http.StatusText(http.StatusNotFound), notfound.Error())
}

func Test_Error_002(t *testing.T) {
	assert := assert.New(t)

	recorder := httptest.NewRecorder()
	err := Error(recorder, Err(http.StatusNotFound))
	assert.NoError(err)
	assert.Equal(http.StatusNotFound, recorder.Code)
}

func Test_Error_003(t *testing.T) {
	assert := assert.New(t)

	recorder := httptest.NewRecorder()
	err := Error(recorder, Err(http.StatusOK), "detail")
	assert.NoError(err)
	assert.Equal(http.StatusOK, recorder.Code)

	expected := ErrResponse{
		Type:   "error",
		Code:   http.StatusOK,
		Reason: http.StatusText(http.StatusOK),
		Detail: "detail",
	}
	var actual ErrResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &actual)
	assert.NoError(err)
	assert.Equal(expected, actual)
}

func Test_Error_004(t *testing.T) {
	assert := assert.New(t)

	recorder := httptest.NewRecorder()
	err := Error(recorder, Err(http.StatusOK), "detail", "detail")
	assert.NoError(err)
	assert.Equal(http.StatusOK, recorder.Code)

	expected := ErrResponse{
		Type:   "error",
		Code:   http.StatusOK,
		Reason: http.StatusText(http.StatusOK),
		Detail: []any{"detail", "detail"},
	}
	var actual ErrResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &actual)
	assert.NoError(err)
	assert.Equal(expected, actual)
}

func Test_Error_With(t *testing.T) {
	assert := assert.New(t)

	t.Run("With", func(t *testing.T) {
		err := ErrBadRequest.With("something went wrong")
		assert.Error(err)
		assert.ErrorIs(err, ErrBadRequest)
		assert.Contains(err.Error(), "something went wrong")
	})

	t.Run("Withf", func(t *testing.T) {
		err := ErrNotFound.Withf("resource %q not found", "abc")
		assert.Error(err)
		assert.ErrorIs(err, ErrNotFound)
		assert.Contains(err.Error(), `resource "abc" not found`)
	})

	t.Run("WithfWrap", func(t *testing.T) {
		cause := &fs.PathError{Op: "open", Path: "file.txt", Err: fs.ErrNotExist}
		err := ErrNotFound.Withf("%w", cause)
		assert.Equal("Not Found: open file.txt: file does not exist", err.Error())
		assert.ErrorIs(err, ErrNotFound)
		assert.ErrorIs(err, fs.ErrNotExist)

		var pathErr *fs.PathError
		assert.ErrorAs(err, &pathErr)
		assert.Equal("file.txt", pathErr.Path)

		var code Err
		assert.ErrorAs(err, &code)
		assert.Equal(ErrNotFound, code)
	})

	t.Run("WithfWrapWithMessage", func(t *testing.T) {
		cause := errors.New("cause")
		err := ErrConflict.Withf("object %q: %w", "abc", cause)
		assert.Equal(`Conflict: object "abc": cause`, err.Error())
		assert.ErrorIs(err, ErrConflict)
		assert.ErrorIs(err, cause)
	})

	t.Run("WithfWrapWritesStatus", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		err := Error(recorder, ErrNotFound.Withf("%w", fs.ErrNotExist))
		assert.NoError(err)
		assert.Equal(http.StatusNotFound, recorder.Code)

		var actual ErrResponse
		assert.NoError(json.Unmarshal(recorder.Body.Bytes(), &actual))
		assert.Equal("Not Found: file does not exist", actual.Reason)
	})

	t.Run("ErrorWritesWrappedErr", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		plainErr := ErrBadRequest.With("wrapped")
		writeErr := Error(recorder, plainErr)
		assert.NoError(writeErr)
		assert.Equal(http.StatusBadRequest, recorder.Code)

		var actual ErrResponse
		err := json.Unmarshal(recorder.Body.Bytes(), &actual)
		assert.NoError(err)
		assert.Equal(http.StatusBadRequest, actual.Code)
		assert.Contains(actual.Reason, "wrapped")
	})
}
