package logger_test

import (
	"fmt"
	"log/slog"
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
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	ts := httptest.NewServer(logger.NewMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	})))
	defer ts.Close()

	response, err := http.Get(ts.URL)
	if !assert.NoError(err) {
		assert.FailNow("Failed to get URL")
	}
	assert.Equal(http.StatusOK, response.StatusCode)
}
