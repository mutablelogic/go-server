package httpresponse

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// TextStream implements a stream of text events
type TextStream struct {
	wg  sync.WaitGroup
	w   io.Writer
	ch  chan *textevent
	err error
}

type textevent struct {
	name string
	data []any
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultKeepAlive = 10 * time.Second
)

var (
	strPing    = "ping"
	strEvent   = []byte("event: ")
	strData    = []byte("data: ")
	strNewline = []byte("\n")
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new text stream with mimetype text/event-stream
// Additional header tuples can be provided as a series of key-value pairs
func NewTextStream(w http.ResponseWriter, tuples ...string) *TextStream {
	// Check parameters
	if w == nil {
		return nil
	}
	if len(tuples)%2 != 0 {
		return nil
	}

	// Create a text stream
	self := new(TextStream)
	self.w = w
	self.ch = make(chan *textevent)

	// Set the default content type
	w.Header().Set(ContentTypeKey, ContentTypeTextStream)

	// Set additional headers
	for i := 0; i < len(tuples); i += 2 {
		w.Header().Set(tuples[i], tuples[i+1])
	}

	// Write the response, don't know is this is the right one
	w.WriteHeader(http.StatusContinue)

	// goroutine will write to the response writer until the channel is closed
	self.wg.Add(1)
	go func() {
		defer self.wg.Done()

		// Create a ticker for ping messages
		ticker := time.NewTimer(100 * time.Millisecond)
		defer ticker.Stop()

		// Run until the channel is closed
		for {
			select {
			case evt := <-self.ch:
				if evt == nil {
					return
				}
				self.emit(evt)
				ticker.Reset(defaultKeepAlive)
			case <-ticker.C:
				self.err = errors.Join(self.err, self.emit(&textevent{strPing, nil}))
				ticker.Reset(defaultKeepAlive)
			}
		}
	}()

	// Return the textstream object
	return self
}

// Close the text stream to stop sending ping messages
func (s *TextStream) Close() error {
	// Close the channel
	close(s.ch)

	// Wait for the goroutine to finish
	s.wg.Wait()

	// Return any errors
	return s.err
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Write a text event to the stream, and one or more optional data objects
// which are encoded as JSON
func (s *TextStream) Write(name string, data ...any) {
	s.ch <- &textevent{name, data}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// emit an event to the stream
func (s *TextStream) emit(e *textevent) error {
	var result error

	// Write the event to the stream
	if e.name != "" {
		if err := s.write(strEvent, []byte(e.name), strNewline); err != nil {
			return err
		}
	}

	// Write the data to the stream
	for _, v := range e.data {
		if v == nil {
			continue
		} else if data, err := json.Marshal(v); err != nil {
			result = errors.Join(result, err)
		} else if err := s.write(strData, data, strNewline); err != nil {
			result = errors.Join(result, err)
		}
	}

	// Flush the event
	if result == nil {
		if err := s.write(strNewline); err != nil {
			result = errors.Join(result, err)
		}
		if w, ok := s.w.(http.Flusher); ok {
			w.Flush()
		}
	}

	// Return any errors
	return result
}

func (s *TextStream) write(v ...[]byte) error {
	if _, err := s.w.Write(bytes.Join(v, nil)); err != nil {
		return err
	}
	return nil
}
