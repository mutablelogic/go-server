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
	mu      sync.Mutex
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

// SetTimeout changes the timeout of each write, and the deadline of the
// write in progress, to the timeout from now, such as to end a stream to a
// client which stopped reading sooner, when the server is stopping. It can
// be called while a write is in progress.
func (w *DeadlineWriter) SetTimeout(timeout time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.timeout = timeout
	w.setDeadline()
}

// Unwrap returns the response writer, for a response controller
func (w *DeadlineWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// deadline sets the deadline of the next write, of the timeout, which is
// read and applied under the lock, so a write can't apply a timeout which
// SetTimeout has changed since it was read
func (w *DeadlineWriter) deadline() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.setDeadline()
}

// setDeadline sets the deadline of the timeout from now, which must be
// locked
func (w *DeadlineWriter) setDeadline() {
	if err := w.rc.SetWriteDeadline(time.Now().Add(w.timeout)); err != nil && !errors.Is(err, http.ErrNotSupported) {
		w.fail()
	}
}

// fail closes the channel which says a write failed
func (w *DeadlineWriter) fail() {
	w.once.Do(func() { close(w.failed) })
}
