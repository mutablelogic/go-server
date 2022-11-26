package expr

import "fmt"

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Pos marks the position within a context, like a file path. If path is
// nil then Pos is not used.
type Pos struct {
	Path *string
	Line uint
	Col  uint
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *Pos) String() string {
	if p.Path == nil || *p.Path == "" {
		return fmt.Sprintf("pos<@%d:%d>", p.Line, p.Col)
	} else {
		return fmt.Sprintf("pos<%s@%d:%d>", *p.Path, p.Line, p.Col)
	}
}
