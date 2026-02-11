package httpresponse

import (
	"encoding/json"
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

	expected := errjson{
		Code:   http.StatusOK,
		Reason: http.StatusText(http.StatusOK),
		Detail: "detail",
	}
	var actual errjson
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

	expected := errjson{
		Code:   http.StatusOK,
		Reason: http.StatusText(http.StatusOK),
		Detail: []any{"detail", "detail"},
	}
	var actual errjson
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

	t.Run("ErrorWritesWrappedErr", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		plainErr := ErrBadRequest.With("wrapped")
		writeErr := Error(recorder, plainErr)
		assert.NoError(writeErr)
		assert.Equal(http.StatusBadRequest, recorder.Code)

		var actual errjson
		err := json.Unmarshal(recorder.Body.Bytes(), &actual)
		assert.NoError(err)
		assert.Equal(http.StatusBadRequest, actual.Code)
		assert.Contains(actual.Reason, "wrapped")
	})
}
