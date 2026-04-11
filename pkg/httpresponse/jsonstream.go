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

// JSONStream mirrors the client-side streaming interface: Recv yields newline-
// delimited JSON frames from the request body, and Send writes one JSON frame
// to the response body followed by a newline and an immediate flush.
type JSONStream interface {
	Recv() <-chan json.RawMessage
	Send(json.RawMessage) error
}

// JSONStreamConn is the richer stream connection returned by NewJSONStream.
// It adds access to the request context and explicit close semantics.
type JSONStreamConn interface {
	JSONStream
	Context() context.Context
	Close() error
}

type jsonstream struct {
	req       *http.Request
	reader    *bufio.Reader
	w         http.ResponseWriter
	recvMu    sync.Mutex
	mu        sync.Mutex
	recvOnce  sync.Once
	recvCh    chan json.RawMessage
	closeOnce sync.Once
	closed    atomic.Bool
	err       error
}

type JSONStreamHandlerFunc func(req *http.Request, stream JSONStream) error

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new full-duplex JSON stream with mimetype application/ndjson.
// Additional header tuples can be provided as a series of key-value pairs.
func NewJSONStream(w http.ResponseWriter, r *http.Request, headers ...string) (JSONStreamConn, error) {
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
	if rc := http.NewResponseController(w); rc != nil {
		_ = rc.EnableFullDuplex()
	}

	// Create the JSONStream
	self := &jsonstream{
		req:    r,
		reader: bufio.NewReader(r.Body),
		w:      w,
		recvCh: make(chan json.RawMessage),
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
func (s *jsonstream) Context() context.Context {
	return s.req.Context()
}

// Recv returns a channel of newline-delimited JSON frames from the request body.
// Blank lines are treated as keep-alive heartbeats and emitted as nil frames.
func (s *jsonstream) Recv() <-chan json.RawMessage {
	s.recvOnce.Do(func() {
		if s.closed.Load() {
			close(s.recvCh)
			return
		}
		go s.recvLoop()
	})
	return s.recvCh
}

func (s *jsonstream) recvLoop() {
	defer close(s.recvCh)

	for {
		frame, ok := s.recvFrame()
		if !ok {
			return
		}

		select {
		case <-s.Context().Done():
			return
		case s.recvCh <- frame:
		}
	}
}

func (s *jsonstream) recvFrame() (json.RawMessage, bool) {
	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	// Return if already closed
	if s.closed.Load() {
		return nil, false
	}

	// Read the next line from the request body
	line, err := s.reader.ReadBytes('\n')
	if err != nil {
		if err != io.EOF || len(line) == 0 {
			return nil, false
		}
	}

	// Treat a blank line as a keep-alive heartbeat.
	frame := bytes.TrimSpace(line)
	if len(frame) == 0 {
		return nil, true
	}

	// Decode the JSON frame to ensure it's valid JSON and compact it to remove whitespace.
	var raw json.RawMessage
	if err := json.Unmarshal(frame, &raw); err != nil {
		return nil, false
	}

	// Return the compacted JSON frame
	return raw, true
}

// Send writes one JSON frame to the response body and flushes it immediately.
func (s *jsonstream) Send(frame json.RawMessage) error {
	return s.sendFrame(frame)
}

// NewJSONStreamHandler wraps a channel-based handler function in an HTTP handler.
// The receive channel yields compacted JSON frames from the request body and nil
// for heartbeat lines. Sending nil on the write channel emits a blank heartbeat line.
func NewJSONStreamHandler(fn JSONStreamHandlerFunc, headers ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fn == nil {
			_ = Error(w, ErrInternalError.With("json stream handler is nil"))
			return
		}

		stream, err := NewJSONStream(w, r, headers...)
		if err != nil {
			_ = Error(w, err)
			return
		}
		defer stream.Close()

		_ = fn(r, stream)
	})
}

// Close closes the request body and marks the stream as closed.
func (s *jsonstream) Close() error {
	s.closeOnce.Do(func() {
		s.closed.Store(true)
		if s.req != nil && s.req.Body != nil {
			s.err = s.req.Body.Close()
		}
	})
	return s.err
}

func (s *jsonstream) sendFrame(frame json.RawMessage) error {
	if s.closed.Load() {
		return io.ErrClosedPipe
	}

	var data []byte
	if frame == nil {
		data = []byte{'\n'}
	} else {
		var buf bytes.Buffer
		if err := json.Compact(&buf, frame); err != nil {
			return err
		}
		data = append(buf.Bytes(), '\n')
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed.Load() {
		return io.ErrClosedPipe
	}
	if _, err := s.w.Write(data); err != nil {
		return err
	}
	s.w.(http.Flusher).Flush()

	return nil
}
