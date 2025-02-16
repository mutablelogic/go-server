package httpresponse

import (
	"net/http"
	"net/http/httptest"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"
)

func Test_JSON_001(t *testing.T) {
	assert := assert.New(t)

	recorder := httptest.NewRecorder()
	err := JSON(recorder, http.StatusOK, 0, nil)
	assert.NoError(err)
	assert.Equal("application/json", recorder.Header().Get("Content-Type"))
	assert.Equal(http.StatusOK, recorder.Code)
	assert.Equal("null\n", recorder.Body.String())
}

func Test_JSON_002(t *testing.T) {
	assert := assert.New(t)

	recorder := httptest.NewRecorder()
	err := JSON(recorder, http.StatusOK, 1, nil)
	assert.NoError(err)
	assert.Equal("application/json", recorder.Header().Get("Content-Type"))
	assert.Equal(http.StatusOK, recorder.Code)
	assert.Equal("null\n", recorder.Body.String())
}

func Test_JSON_003(t *testing.T) {
	assert := assert.New(t)

	recorder := httptest.NewRecorder()
	err := JSON(recorder, http.StatusOK, 0, []int{1, 2, 3})
	assert.NoError(err)
	assert.Equal("application/json", recorder.Header().Get("Content-Type"))
	assert.Equal(http.StatusOK, recorder.Code)
	assert.Equal("[1,2,3]\n", recorder.Body.String())
}

func Test_JSON_004(t *testing.T) {
	assert := assert.New(t)

	recorder := httptest.NewRecorder()
	err := JSON(recorder, http.StatusOK, 1, []int{1, 2, 3})
	assert.NoError(err)
	assert.Equal("application/json", recorder.Header().Get("Content-Type"))
	assert.Equal(http.StatusOK, recorder.Code)
	assert.Equal("[\n 1,\n 2,\n 3\n]\n", recorder.Body.String())
}
