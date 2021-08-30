package mdns

import (
	"fmt"
	"net"
	"strings"

	// Modules
	. "github.com/djthorpe/go-server"
	dns "github.com/miekg/dns"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type message struct {
	*dns.Msg
	Addr  net.Addr
	Index int
	Zone  string
	Err   error
}

type event struct {
	EventType
	service
}

type EventType int

type ServiceEvent interface {
	Event
	Type() EventType
	Service() Service
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	EVENT_TYPE_ADDED EventType = (1 << iota)
	EVENT_TYPE_REMOVED
	EVENT_TYPE_EXPIRED
	EVENT_TYPE_CHANGED
	EVENT_TYPE_SERVICE
	EVENT_TYPE_NONE EventType = 0
	EVENT_TYPE_MIN            = EVENT_TYPE_ADDED
	EVENT_TYPE_MAX            = EVENT_TYPE_SERVICE
)

///////////////////////////////////////////////////////////////////////////////
// EVENT IMPLEMENTATION

func (e event) Name() string {
	return e.service.Instance()
}

func (e event) Value() interface{} {
	return &e.service
}

func (e event) Type() EventType {
	return e.EventType
}

func (e event) Service() Service {
	return &e.service
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (e event) String() string {
	str := "<mdns.event"
	if t := e.EventType; t != EVENT_TYPE_NONE {
		str += " type=" + e.EventType.String()
	}
	str += fmt.Sprint(" ", e.service)
	return str + ">"
}

func (f EventType) String() string {
	if f == EVENT_TYPE_NONE {
		return f.FlagString()
	}
	str := ""
	for v := EVENT_TYPE_MIN; v <= EVENT_TYPE_MAX; v <<= 1 {
		if f&v == v {
			str += v.FlagString() + "|"
		}
	}
	return strings.TrimSuffix(str, "|")
}

func (v EventType) FlagString() string {
	switch v {
	case EVENT_TYPE_NONE:
		return "EVENT_TYPE_NONE"
	case EVENT_TYPE_ADDED:
		return "EVENT_TYPE_ADDED"
	case EVENT_TYPE_REMOVED:
		return "EVENT_TYPE_REMOVED"
	case EVENT_TYPE_EXPIRED:
		return "EVENT_TYPE_EXPIRED"
	case EVENT_TYPE_CHANGED:
		return "EVENT_TYPE_CHANGED"
	case EVENT_TYPE_SERVICE:
		return "EVENT_TYPE_SERVICE"
	default:
		return "[?? Invalid EventType]"
	}
}
