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

type JSONStreamHandlerFunc func(r <-chan json.RawMessage, w chan<- json.RawMessage) error

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
	if rc := http.NewResponseController(w); rc != nil {
		_ = rc.EnableFullDuplex()
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
	if frame == nil {
		return ErrBadRequest.With("invalid json frame: nil")
	}
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

		recvCh := make(chan json.RawMessage)
		sendCh := make(chan json.RawMessage)

		var wg sync.WaitGroup
		var recvErr error
		var sendErr error

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer close(recvCh)

			for {
				frame, err := stream.Recv()
				if err == io.EOF {
					return
				}
				if err != nil {
					recvErr = err
					return
				}

				select {
				case <-stream.Context().Done():
					return
				case recvCh <- frame:
				}
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			var failed bool
			for frame := range sendCh {
				if failed {
					continue
				}
				if err := stream.sendFrame(frame); err != nil {
					sendErr = err
					failed = true
				}
			}
		}()

		_ = fn(recvCh, sendCh)
		close(sendCh)
		wg.Wait()

		_ = recvErr
		_ = sendErr
	})
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

func (s *JSONStream) sendFrame(frame json.RawMessage) error {
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
