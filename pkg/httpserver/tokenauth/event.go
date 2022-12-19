package tokenauth

/////////////////////////////////////////////////////////////////////
// TYPES

type EventType int

/////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	EventTypeNone EventType = iota
	EventTypeWrite
	EventTypeRotate
)

/////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t EventType) String() string {
	switch t {
	case EventTypeNone:
		return "none"
	case EventTypeWrite:
		return "write"
	case EventTypeRotate:
		return "rotate"
	}
	return "[?? Invalid EventType value]"
}
