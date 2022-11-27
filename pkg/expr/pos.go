package expr

import (
	"fmt"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Pos marks the position within a context, like a file path. If path is
// nil then Pos is not used. Note that Line and Col are zero-indexed, so
// 1 needs to be added to them when printing.
type Pos struct {
	Path *string
	Line uint
	Col  uint

	// Record previous position for unread
	x, y uint
}

// Error with a position
type PosError struct {
	Err error
	Pos Pos
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewPosError(err error, pos Pos) error {
	return &PosError{Err: err, Pos: pos}
}

func NewParseError(t *Token) error {
	return &PosError{Err: ErrBadParameter.Withf("Parse error near %q", t.toString()), Pos: t.Pos}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *Pos) String() string {
	if p.Path == nil || *p.Path == "" {
		return fmt.Sprintf("pos<%d:%d>", p.Line+1, p.Col)
	} else {
		return fmt.Sprintf("pos<%s:%d:%d>", *p.Path, p.Line+1, p.Col+1)
	}
}

func (e *PosError) Error() string {
	if e.Pos.Path == nil || *e.Pos.Path == "" {
		return fmt.Errorf("%d:%d: %w", e.Pos.Line+1, e.Pos.Col, e.Err).Error()
	} else {
		return fmt.Errorf("%s:%d:%d: %w", *e.Pos.Path, e.Pos.Line+1, e.Pos.Col, e.Err).Error()
	}
}
