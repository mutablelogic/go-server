package indexer

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	// Modules
	. "github.com/djthorpe/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type IndexerEvent interface {
	Event
	Key() string
	Type() EventType
	FileInfo() fs.FileInfo
	Path() string
}

type event struct {
	EventType
	name string
	root string
	path string
	info fs.FileInfo
}

type EventType int

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	EVENT_TYPE_ADDED EventType = (1 << iota)
	EVENT_TYPE_REMOVED
	EVENT_TYPE_RENAMED
	EVENT_TYPE_CHANGED
	EVENT_TYPE_NONE EventType = 0
	EVENT_TYPE_MIN            = EVENT_TYPE_ADDED
	EVENT_TYPE_MAX            = EVENT_TYPE_CHANGED
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (this *Indexer) NewEvent(t EventType, abspath string, info fs.FileInfo) (event, error) {
	relpath, err := filepath.Rel(this.path, abspath)
	if err != nil {
		return event{}, err
	}
	return event{
		EventType: t,
		name:      this.name,
		root:      this.path,
		path:      relpath,
		info:      info,
	}, nil
}

///////////////////////////////////////////////////////////////////////////////
// EVENT IMPLEMENTATION

func (e event) Name() string {
	return e.name
}

func (e event) Value() interface{} {
	return e.path
}

func (e event) Type() EventType {
	return e.EventType
}

func (e event) Key() string {
	return filepath.Join(e.root, e.path)
}

func (e event) FileInfo() fs.FileInfo {
	return e.info
}

func (e event) Path() string {
	return e.path
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (e event) String() string {
	str := "<indexer.event"
	if t := e.EventType; t != EVENT_TYPE_NONE {
		str += " type=" + e.EventType.String()
	}
	if name := e.Name(); name != "" {
		str += fmt.Sprintf(" name=%q", name)
	}
	if path := e.Value(); path != nil {
		str += fmt.Sprintf(" path=%q", path)
	}
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
	case EVENT_TYPE_RENAMED:
		return "EVENT_TYPE_RENAMED"
	case EVENT_TYPE_CHANGED:
		return "EVENT_TYPE_CHANGED"
	default:
		return "[?? Invalid EventType]"
	}
}
