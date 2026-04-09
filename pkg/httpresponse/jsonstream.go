package httpresponse

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"sync/atomic"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// JSONStream implements a full-duplex newline-delimited JSON stream.
//
// Recv reads exactly one JSON value per line from the request body and Send
// writes one JSON value to the response body followed by a newline and an
// immediate flush.
type JSONStream struct {
	req       *http.Request
	reader    *bufio.Reader
	w         http.ResponseWriter
	recvMu    sync.Mutex
	mu        sync.Mutex
	closeOnce sync.Once
	closed    atomic.Bool
	err       error
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new full-duplex JSON stream with mimetype application/ndjson.
// Additional header tuples can be provided as a series of key-value pairs.
func NewJSONStream(w http.ResponseWriter, r *http.Request, headers ...string) (*JSONStream, error) {
	if w == nil {
		return nil, ErrBadRequest.With("response writer is nil")
	}
	if r == nil {
		return nil, ErrBadRequest.With("request is nil")
	}
	if len(headers)%2 != 0 {
		return nil, ErrBadRequest.With("headers must be key/value pairs")
	}
	if _, ok := w.(http.Flusher); !ok {
		return nil, ErrInternalError.With("response writer does not support streaming")
	}

	// Create the JSONStream
	self := &JSONStream{
		req:    r,
		reader: bufio.NewReader(r.Body),
		w:      w,
	}

	// Set the headers on the response, and write out the header
	w.Header().Set(types.ContentTypeHeader, types.ContentTypeJSONStream)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	for i := 0; i < len(headers); i += 2 {
		w.Header().Set(headers[i], headers[i+1])
	}
	w.WriteHeader(http.StatusOK)
	w.(http.Flusher).Flush()

	// Return success
	return self, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Context returns the request context associated with the stream.
func (s *JSONStream) Context() context.Context {
	return s.req.Context()
}

// Recv returns the next newline-delimited JSON frame from the request body.
// A blank line is treated as a keep-alive heartbeat and returns nil, nil.
func (s *JSONStream) Recv() (json.RawMessage, error) {
	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	// Return if already closed
	if s.closed.Load() {
		return nil, io.ErrClosedPipe
	}

	// Read the next line from the request body
	line, err := s.reader.ReadBytes('\n')
	if err != nil {
		if err != io.EOF || len(line) == 0 {
			return nil, err
		}
	}

	// Treat a blank line as a keep-alive heartbeat.
	frame := bytes.TrimSpace(line)
	if len(frame) == 0 {
		return nil, nil
	}

	// Decode the JSON frame to ensure it's valid JSON and compact it to remove whitespace.
	var raw json.RawMessage
	if err := json.Unmarshal(frame, &raw); err != nil {
		return nil, ErrBadRequest.Withf("invalid json frame: %v", err)
	}

	// Return the compacted JSON frame
	return raw, nil
}

// Send writes one JSON frame to the response body and flushes it immediately.
func (s *JSONStream) Send(frame json.RawMessage) error {
	if s.closed.Load() {
		return io.ErrClosedPipe
	}

	var buf bytes.Buffer
	if err := json.Compact(&buf, frame); err != nil {
		return err
	}
	data := buf.Bytes()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed.Load() {
		return io.ErrClosedPipe
	}
	if _, err := s.w.Write(data); err != nil {
		return err
	}
	if _, err := s.w.Write([]byte{'\n'}); err != nil {
		return err
	}
	s.w.(http.Flusher).Flush()

	return nil
}

// Close closes the request body and marks the stream as closed.
func (s *JSONStream) Close() error {
	s.closeOnce.Do(func() {
		s.closed.Store(true)
		if s.req != nil && s.req.Body != nil {
			s.err = s.req.Body.Close()
		}
	})
	return s.err
}
