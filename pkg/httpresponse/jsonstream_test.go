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

func recvFrame(t *testing.T, stream httpresponse.JSONStream) (json.RawMessage, bool) {
	t.Helper()
	ch := stream.Recv()
	require.NotNil(t, ch)
	frame, ok := <-ch
	return frame, ok
}

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

	frame, ok := recvFrame(t, stream)
	require.True(t, ok)
	assert.Equal(json.RawMessage(`{"a":1}`), frame)

	frame, ok = recvFrame(t, stream)
	require.True(t, ok)
	assert.Equal(json.RawMessage(`{"b":2}`), frame)

	frame, ok = recvFrame(t, stream)
	assert.Nil(frame)
	assert.False(ok)
}

func Test_jsonstream_recv_blank_line_is_keepalive(t *testing.T) {
	assert := assert.New(t)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{\"a\":1}\n\n{\"b\":2}\n"))
	resp := httptest.NewRecorder()

	stream, err := httpresponse.NewJSONStream(resp, req)
	require.NoError(t, err)
	defer stream.Close()

	frame, ok := recvFrame(t, stream)
	require.True(t, ok)
	assert.Equal(json.RawMessage(`{"a":1}`), frame)

	frame, ok = recvFrame(t, stream)
	assert.Nil(frame)
	assert.True(ok)

	frame, ok = recvFrame(t, stream)
	require.True(t, ok)
	assert.Equal(json.RawMessage(`{"b":2}`), frame)
}

func Test_jsonstream_recv_trailing_blank_line_is_keepalive(t *testing.T) {
	assert := assert.New(t)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{\"a\":1}\n\n"))
	resp := httptest.NewRecorder()

	stream, err := httpresponse.NewJSONStream(resp, req)
	require.NoError(t, err)
	defer stream.Close()

	frame, ok := recvFrame(t, stream)
	require.True(t, ok)
	assert.Equal(json.RawMessage(`{"a":1}`), frame)

	frame, ok = recvFrame(t, stream)
	assert.Nil(frame)
	assert.True(ok)

	frame, ok = recvFrame(t, stream)
	assert.Nil(frame)
	assert.False(ok)

	frame, ok = recvFrame(t, stream)
	assert.Nil(frame)
	assert.False(ok)
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

	assert.NoError(stream.Send(nil))
	assert.Error(stream.Send(json.RawMessage("not-json")))

	require.NoError(t, stream.Close())
	assert.ErrorIs(stream.Send(json.RawMessage(`{"ok":true}`)), io.ErrClosedPipe)
	frame, ok := recvFrame(t, stream)
	assert.Nil(frame)
	assert.False(ok)
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
	frame, ok := recvFrame(t, stream)
	assert.Nil(frame)
	assert.False(ok)
}

func Test_jsonstream_handler_echo(t *testing.T) {
	srv := httptest.NewServer(httpresponse.NewJSONStreamHandler(func(req *http.Request, stream httpresponse.JSONStream) error {
		assert.Equal(t, http.MethodPost, req.Method)
		for frame := range stream.Recv() {
			if frame == nil {
				continue
			}
			if err := stream.Send(frame); err != nil {
				return err
			}
		}
		return nil
	}, "X-Test", "ok"))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodPost, srv.URL, strings.NewReader("{\"a\":1}\n{\"b\":2}\n"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/ndjson")
	req.Header.Set("Accept", "application/ndjson")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/ndjson", resp.Header.Get("Content-Type"))
	assert.Equal(t, "ok", resp.Header.Get("X-Test"))
	assert.Equal(t, "{\"a\":1}\n{\"b\":2}\n", string(data))
}

func Test_jsonstream_handler_heartbeat_forwarding(t *testing.T) {
	srv := httptest.NewServer(httpresponse.NewJSONStreamHandler(func(req *http.Request, stream httpresponse.JSONStream) error {
		assert.Equal(t, "/", req.URL.Path)
		for frame := range stream.Recv() {
			if err := stream.Send(frame); err != nil {
				return err
			}
		}
		return nil
	}))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodPost, srv.URL, strings.NewReader("\n{\"ok\":true}\n"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/ndjson")
	req.Header.Set("Accept", "application/ndjson")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, "\n{\"ok\":true}\n", string(data))
}

func Test_jsonstream_handler_nil(t *testing.T) {
	srv := httptest.NewServer(httpresponse.NewJSONStreamHandler(nil))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodPost, srv.URL, http.NoBody)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(data), "json stream handler is nil")
}
