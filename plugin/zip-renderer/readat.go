package main

import (
	"io"
	"io/ioutil"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ReaderAt struct {
	r io.Reader
	n int64
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewReaderAt(r io.Reader) io.ReaderAt {
	return &ReaderAt{r, 0}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (r *ReaderAt) ReadAt(p []byte, off int64) (int, error) {
	if off < r.n {
		return 0, ErrOutOfOrder.Withf("invalid offset, %v < %v", off, r.n)
	}
	diff := off - r.n
	written, err := io.CopyN(ioutil.Discard, r.r, diff)
	r.n += written
	if err != nil {
		return 0, err
	}

	n, err := r.r.Read(p)
	r.n += int64(n)
	return n, err
}
