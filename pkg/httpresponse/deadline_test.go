package httpresponse_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

func Test_DeadlineWriter(t *testing.T) {
	t.Run("Write", func(t *testing.T) {
		// A stream is written and flushed to a client which reads it, and
		// doesn't fail
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			dw, err := httpresponse.NewDeadlineWriter(w, time.Second)
			if !assert.NoError(t, err) {
				return
			}
			for _, line := range []string{"one\n", "two\n"} {
				_, err := dw.Write([]byte(line))
				assert.NoError(t, err)
				dw.Flush()
			}
			select {
			case <-dw.Failed():
				t.Error("the writer failed")
			default:
			}
		}))
		defer server.Close()

		resp, err := http.Get(server.URL)
		require.NoError(t, err)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, "one\ntwo\n", string(body))
	})

	t.Run("Unsupported", func(t *testing.T) {
		// A response writer which can't have a deadline is written to, and
		// is unwrapped for a response controller
		w := httptest.NewRecorder()
		dw, err := httpresponse.NewDeadlineWriter(w, time.Second)
		require.NoError(t, err)
		_, err = dw.Write([]byte("data"))
		require.NoError(t, err)
		assert.Equal(t, http.ResponseWriter(w), dw.Unwrap())
		assert.True(t, strings.HasPrefix(w.Body.String(), "data"))
	})

	t.Run("ClientGone", func(t *testing.T) {
		// A write to a client which has gone away fails, and Failed is closed
		failed := make(chan bool, 1)
		closed := make(chan struct{})
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			dw, err := httpresponse.NewDeadlineWriter(w, 100*time.Millisecond)
			if !assert.NoError(t, err) {
				return
			}
			dw.Write([]byte("first\n"))
			dw.Flush()
			<-closed
			deadline := time.Now().Add(5 * time.Second)
			for time.Now().Before(deadline) {
				dw.Write([]byte(strings.Repeat("x", 64*1024)))
				dw.Flush()
				select {
				case <-dw.Failed():
					failed <- true
					return
				default:
				}
			}
			failed <- false
		}))
		defer server.Close()

		resp, err := http.Get(server.URL)
		require.NoError(t, err)
		resp.Body.Close()
		close(closed)
		assert.True(t, <-failed)
	})
}
