package types

import (
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Bool bool
type Int int64
type UInt uint64
type Float float64
type Duration time.Duration
type String string
type Eval string // Evaluation into a different type
