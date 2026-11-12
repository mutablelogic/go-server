package httpresponse

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// DeadlineWriter writes a stream to a client, such as of server-sent events,
// with a deadline for each write, rather than the server's write timeout,
// which is for the whole response, so a client which stops reading can't
// block it for longer than the timeout, and its Failed channel is closed
// when a write fails
type DeadlineWriter struct {
	http.ResponseWriter
	rc      *http.ResponseController
	timeout time.Duration
	once    sync.Once
	failed  chan struct{}
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewDeadlineWriter returns a writer of a stream to a client, with the
// timeout of each write, and clears the server's deadline of the whole
// response, unless the response writer can't have its deadline set, when
// it's streamed to until the server's timeout
func NewDeadlineWriter(w http.ResponseWriter, timeout time.Duration) (*DeadlineWriter, error) {
	self := &DeadlineWriter{ResponseWriter: w, rc: http.NewResponseController(w), timeout: timeout, failed: make(chan struct{})}
	if err := self.rc.SetWriteDeadline(time.Time{}); err != nil && !errors.Is(err, http.ErrNotSupported) {
		return nil, err
	}
	return self, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Failed returns a channel which is closed when a write fails, as the client
// stopped reading, or went away
func (w *DeadlineWriter) Failed() <-chan struct{} {
	return w.failed
}

// Write writes to the client, by the deadline
func (w *DeadlineWriter) Write(data []byte) (int, error) {
	w.deadline()
	n, err := w.ResponseWriter.Write(data)
	if err != nil {
		w.fail()
	}
	return n, err
}

// Flush sends what's written to the client, by the deadline
func (w *DeadlineWriter) Flush() {
	w.deadline()
	if err := w.rc.Flush(); err != nil {
		w.fail()
	}
}

// Unwrap returns the response writer, for a response controller
func (w *DeadlineWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// deadline sets the deadline of the next write
func (w *DeadlineWriter) deadline() {
	if err := w.rc.SetWriteDeadline(time.Now().Add(w.timeout)); err != nil && !errors.Is(err, http.ErrNotSupported) {
		w.fail()
	}
}

// fail closes the channel which says a write failed
func (w *DeadlineWriter) fail() {
	w.once.Do(func() { close(w.failed) })
}
