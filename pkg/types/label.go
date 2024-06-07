package types

import "strings"

/////////////////////////////////////////////////////////////////////
// TYPES

type Label string

/////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	LabelSeparator = "."
)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// NewLabel returns a new label which is a concatenation of the prefix
// and parts, separated by a period. Returns an empty string if the
// prefix or any of the parts are not valid identifiers
func NewLabel(prefix string, parts ...string) Label {
	if !IsIdentifier(prefix) {
		return ""
	}
	for _, part := range parts {
		if !IsIdentifier(part) {
			return ""
		}
	}
	if len(parts) == 0 {
		return Label(prefix)
	} else {
		return Label(prefix + LabelSeparator + strings.Join(parts, LabelSeparator))
	}
}
