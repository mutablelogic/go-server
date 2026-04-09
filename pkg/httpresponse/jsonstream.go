package httpresponse

import (
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
// Recv decodes one JSON value from the request body, and Send writes one JSON
// value to the response body followed by a newline and an immediate flush.
type JSONStream struct {
	ctx       context.Context
	req       *http.Request
	dec       *json.Decoder
	w         http.ResponseWriter
	flusher   http.Flusher
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

	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, ErrInternalError.With("response writer does not support streaming")
	}

	body := r.Body
	if body == nil {
		body = http.NoBody
	}

	self := &JSONStream{
		ctx:     r.Context(),
		req:     r,
		dec:     json.NewDecoder(body),
		w:       w,
		flusher: flusher,
	}

	w.Header().Set(types.ContentTypeHeader, types.ContentTypeJSONStream)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	for i := 0; i < len(headers); i += 2 {
		w.Header().Set(headers[i], headers[i+1])
	}

	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	return self, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Context returns the request context associated with the stream.
func (s *JSONStream) Context() context.Context {
	if s == nil || s.ctx == nil {
		return context.Background()
	}
	return s.ctx
}

// Recv returns the next JSON frame from the request body.
func (s *JSONStream) Recv() (json.RawMessage, error) {
	if s == nil || s.closed.Load() {
		return nil, io.ErrClosedPipe
	}

	var frame json.RawMessage
	if err := s.dec.Decode(&frame); err != nil {
		return nil, err
	}

	return frame, nil
}

// Send writes one JSON frame to the response body and flushes it immediately.
func (s *JSONStream) Send(frame json.RawMessage) error {
	if s == nil || s.closed.Load() {
		return io.ErrClosedPipe
	}

	data, err := compactJSONFrame(frame)
	if err != nil {
		return err
	}

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
	s.flusher.Flush()

	return nil
}

// Close closes the request body and marks the stream as closed.
func (s *JSONStream) Close() error {
	if s == nil {
		return nil
	}

	s.closeOnce.Do(func() {
		s.closed.Store(true)
		if s.req != nil && s.req.Body != nil {
			s.err = s.req.Body.Close()
		}
	})

	return s.err
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func compactJSONFrame(frame json.RawMessage) ([]byte, error) {
	frame = bytes.TrimSpace(frame)
	if len(frame) == 0 {
		return nil, ErrBadRequest.With("json frame is empty")
	}

	var buf bytes.Buffer
	if err := json.Compact(&buf, frame); err != nil {
		return nil, ErrBadRequest.Withf("invalid json frame: %v", err)
	}

	return buf.Bytes(), nil
}
