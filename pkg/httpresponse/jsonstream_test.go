package httpresponse_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type trackedReadCloser struct {
	reader io.Reader
	closed int
}

func (t *trackedReadCloser) Read(p []byte) (int, error) {
	return t.reader.Read(p)
}

func (t *trackedReadCloser) Close() error {
	t.closed++
	return nil
}

func Test_jsonstream_new(t *testing.T) {
	assert := assert.New(t)

	req := httptest.NewRequest(http.MethodPost, "/", http.NoBody)
	resp := httptest.NewRecorder()

	stream, err := httpresponse.NewJSONStream(resp, req, "X-Test", "ok")
	require.NoError(t, err)
	require.NotNil(t, stream)

	assert.Equal(http.StatusOK, resp.Code)
	assert.Equal("application/ndjson", resp.Header().Get("Content-Type"))
	assert.Equal("no-cache", resp.Header().Get("Cache-Control"))
	assert.Equal("keep-alive", resp.Header().Get("Connection"))
	assert.Equal("no", resp.Header().Get("X-Accel-Buffering"))
	assert.Equal("ok", resp.Header().Get("X-Test"))
	assert.NotNil(stream.Context())
	assert.NoError(stream.Close())
}

func Test_jsonstream_new_errors(t *testing.T) {
	assert := assert.New(t)

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", http.NoBody)

	stream, err := httpresponse.NewJSONStream(nil, req)
	assert.Nil(stream)
	assert.Error(err)

	stream, err = httpresponse.NewJSONStream(resp, nil)
	assert.Nil(stream)
	assert.Error(err)

	stream, err = httpresponse.NewJSONStream(resp, req, "X-Test")
	assert.Nil(stream)
	assert.Error(err)
}

func Test_jsonstream_recv(t *testing.T) {
	assert := assert.New(t)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{\"a\":1}\n{\"b\":2}\n"))
	resp := httptest.NewRecorder()

	stream, err := httpresponse.NewJSONStream(resp, req)
	require.NoError(t, err)
	defer stream.Close()

	frame, err := stream.Recv()
	require.NoError(t, err)
	assert.Equal(json.RawMessage(`{"a":1}`), frame)

	frame, err = stream.Recv()
	require.NoError(t, err)
	assert.Equal(json.RawMessage(`{"b":2}`), frame)

	frame, err = stream.Recv()
	assert.Nil(frame)
	assert.ErrorIs(err, io.EOF)
}

func Test_jsonstream_send(t *testing.T) {
	assert := assert.New(t)

	req := httptest.NewRequest(http.MethodPost, "/", http.NoBody)
	resp := httptest.NewRecorder()

	stream, err := httpresponse.NewJSONStream(resp, req)
	require.NoError(t, err)
	defer stream.Close()

	err = stream.Send(json.RawMessage(" { \"a\": 1, \"b\": [ 1, 2 ] } "))
	require.NoError(t, err)
	err = stream.Send(json.RawMessage("\n{\n  \"ok\": true\n}\n"))
	require.NoError(t, err)

	assert.Equal("{\"a\":1,\"b\":[1,2]}\n{\"ok\":true}\n", resp.Body.String())
	assert.Equal("application/ndjson", resp.Header().Get("Content-Type"))
}

func Test_jsonstream_send_errors(t *testing.T) {
	assert := assert.New(t)

	req := httptest.NewRequest(http.MethodPost, "/", http.NoBody)
	resp := httptest.NewRecorder()

	stream, err := httpresponse.NewJSONStream(resp, req)
	require.NoError(t, err)

	assert.Error(stream.Send(nil))
	assert.Error(stream.Send(json.RawMessage("not-json")))

	require.NoError(t, stream.Close())
	assert.ErrorIs(stream.Send(json.RawMessage(`{"ok":true}`)), io.ErrClosedPipe)
	frame, err := stream.Recv()
	assert.Nil(frame)
	assert.ErrorIs(err, io.ErrClosedPipe)
}

func Test_jsonstream_close(t *testing.T) {
	assert := assert.New(t)

	body := &trackedReadCloser{reader: strings.NewReader(`{"a":1}`)}
	req := httptest.NewRequest(http.MethodPost, "/", http.NoBody)
	req.Body = body
	resp := httptest.NewRecorder()

	stream, err := httpresponse.NewJSONStream(resp, req)
	require.NoError(t, err)

	assert.NoError(stream.Close())
	assert.NoError(stream.Close())
	assert.Equal(1, body.closed)
	assert.ErrorIs(stream.Send(json.RawMessage(`{"ok":true}`)), io.ErrClosedPipe)
	frame, err := stream.Recv()
	assert.Nil(frame)
	assert.ErrorIs(err, io.ErrClosedPipe)
}
