package httprequest

import (
	"io"
	"mime/multipart"
	"sync"
	"sync/atomic"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// multipartCleanup tracks how many file bodies are still open for a parsed
// multipart form and calls RemoveAll exactly once when the count reaches zero.
// This ensures temp files are removed and in-memory buffers released without
// requiring the caller to manage the form lifetime explicitly.
type multipartCleanup struct {
	form *multipart.Form
	refs atomic.Int32
	once sync.Once
}

// refCountedBody wraps an io.ReadCloser and decrements the multipartCleanup
// ref count on Close; when the count reaches zero RemoveAll is called.
type refCountedBody struct {
	io.ReadCloser
	mc   *multipartCleanup
	once sync.Once
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

// open opens a file part, increments the ref count, and returns a
// refCountedBody whose Close() will decrement the count and trigger cleanup.
func (mc *multipartCleanup) open(fh *multipart.FileHeader) (io.ReadCloser, error) {
	rc, err := fh.Open()
	if err != nil {
		return nil, err
	}
	mc.refs.Add(1)
	return &refCountedBody{ReadCloser: rc, mc: mc}, nil
}

// done must be called after all calls to open have been made. If no files were
// opened (refs == 0), it calls RemoveAll immediately since no closer will do so.
func (mc *multipartCleanup) done() {
	if mc.refs.Load() == 0 {
		mc.removeAll()
	}
}

func (mc *multipartCleanup) removeAll() {
	mc.once.Do(func() { _ = mc.form.RemoveAll() })
}

// Close decrements the ref count; when it reaches zero RemoveAll is called.
func (b *refCountedBody) Close() error {
	err := b.ReadCloser.Close()
	b.once.Do(func() {
		if b.mc.refs.Add(-1) == 0 {
			b.mc.removeAll()
		}
	})
	return err
}
