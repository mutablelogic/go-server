package httpresponse_test

import (
	"net/http/httptest"
	"testing"
	"time"

	// Packages
	"github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/stretchr/testify/assert"
)

func Test_textstream_001(t *testing.T) {
	assert := assert.New(t)

	t.Run("New", func(t *testing.T) {
		resp := httptest.NewRecorder()
		ts := httpresponse.NewTextStream(resp)
		assert.NotNil(ts)
		assert.NoError(ts.Close())
	})

	t.Run("Ping", func(t *testing.T) {
		resp := httptest.NewRecorder()
		ts := httpresponse.NewTextStream(resp)
		assert.NotNil(ts)

		time.Sleep(1 * time.Second)
		assert.NoError(ts.Close())
		assert.Equal(100, resp.Code)
		assert.Equal("text/event-stream", resp.Header().Get("Content-Type"))
		assert.Equal("event: ping\n\n", resp.Body.String())
	})

	t.Run("EventDataAfterPing", func(t *testing.T) {
		resp := httptest.NewRecorder()
		ts := httpresponse.NewTextStream(resp)
		assert.NotNil(ts)

		time.Sleep(200 * time.Millisecond)
		ts.Write("foo")

		time.Sleep(1 * time.Second)
		assert.NoError(ts.Close())
		assert.Equal(100, resp.Code)
		assert.Equal("text/event-stream", resp.Header().Get("Content-Type"))
		assert.Equal("event: ping\n\n"+"event: foo\n\n", resp.Body.String())
	})

	t.Run("EventDataNoPing", func(t *testing.T) {
		resp := httptest.NewRecorder()
		ts := httpresponse.NewTextStream(resp)
		assert.NotNil(ts)

		ts.Write("foo", "bar")

		time.Sleep(1 * time.Second)
		assert.NoError(ts.Close())
		assert.Equal(100, resp.Code)
		assert.Equal("text/event-stream", resp.Header().Get("Content-Type"))
		assert.Equal("event: foo\n"+"data: \"bar\"\n\n", resp.Body.String())
	})

	t.Run("EventDataData", func(t *testing.T) {
		resp := httptest.NewRecorder()
		ts := httpresponse.NewTextStream(resp)
		assert.NotNil(ts)

		ts.Write("foo", "bar1", "bar2")

		time.Sleep(1 * time.Second)
		assert.NoError(ts.Close())
		assert.Equal(100, resp.Code)
		assert.Equal("text/event-stream", resp.Header().Get("Content-Type"))
		assert.Equal("event: foo\n"+"data: \"bar1\"\n"+"data: \"bar2\"\n\n", resp.Body.String())
	})

}
