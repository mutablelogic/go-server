package tui

import (
	"fmt"
	"io"
	"strings"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type summary struct {
	name    string
	count   uint
	bodyLen uint
	offset  uint64
	limit   *uint64
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func TableSummary(name string, count uint, bodyLen uint, offset uint64, limit *uint64) *summary {
	return &summary{
		name:    strings.TrimSpace(name),
		count:   count,
		bodyLen: bodyLen,
		offset:  offset,
		limit:   limit,
	}
}

///////////////////////////////////////////////////////////////////////////////
// RENDER

func (s *summary) Write(w io.Writer) (int, error) {
	return io.WriteString(w, fmt.Sprintln(s.String()))
}

func (s *summary) String() string {
	name := s.name
	if name == "" {
		name = "items"
	}

	// Total count unknown — derive range from body length alone
	if s.count == 0 {
		if s.bodyLen == 0 {
			return fmt.Sprintf("No %s returned", name)
		}
		start := s.offset + 1
		end := s.offset + uint64(s.bodyLen)
		return fmt.Sprintf("Showing %s %d-%d", name, start, end)
	}

	end := uint64(s.count)
	if s.limit != nil && *s.limit > 0 {
		end = min(end, s.offset+*s.limit)
	}
	if s.offset == 0 && end >= uint64(s.count) {
		return fmt.Sprintf("Showing all %d %s", s.count, name)
	}

	start := min(uint64(s.count), s.offset+1)
	end = min(uint64(s.count), end)
	return fmt.Sprintf("Showing %s %d-%d of %d", name, start, end, s.count)
}
