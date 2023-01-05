package mdns

///////////////////////////////////////////////////////////////////////////////
// TYPES

type EventType uint

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	None EventType = iota
	Question
	Answer
	Registered
	Resolved
	Expired
	Sent
)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t EventType) String() string {
	switch t {
	case None:
		return "None"
	case Question:
		return "Question"
	case Answer:
		return "Answer"
	case Registered:
		return "Registered"
	case Resolved:
		return "Resolved"
	case Expired:
		return "Expired"
	case Sent:
		return "Sent"
	}
	return "[?? Invalid EventType value]"
}
