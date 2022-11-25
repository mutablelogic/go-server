package nginx

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Event uint

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	EventStart Event = iota
	EventStop
	EventReload
	EventInfo
	EventError
)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (e Event) String() string {
	switch e {
	case EventStart:
		return "start"
	case EventStop:
		return "stop"
	case EventReload:
		return "reload"
	case EventInfo:
		return "info"
	case EventError:
		return "error"
	default:
		return "[?? Invalid Event value]"
	}
}
