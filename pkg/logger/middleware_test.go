package logger_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	// Packages
	logger "github.com/mutablelogic/go-server/pkg/logger"
	assert "github.com/stretchr/testify/assert"
)

func Test_Middleware_001(t *testing.T) {
	assert := assert.New(t)
	logger := logger.New(os.Stderr, logger.Term, false)
	ts := httptest.NewServer(logger.HandleFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	})))
	defer ts.Close()

	response, err := http.Get(ts.URL)
	if !assert.NoError(err) {
		assert.FailNow("Failed to get URL")
	}
	assert.Equal(http.StatusOK, response.StatusCode)
}
