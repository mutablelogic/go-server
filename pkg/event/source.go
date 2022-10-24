package event

import (
	"fmt"
	"sync"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Source struct {
	sync.RWMutex
	Cap int          // Capacity of any created channels, zero means unbuffered
	ch  []chan Event // subscribed channels
}

// Compile time check
var _ EventSource = (*Source)(nil)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s *Source) String() string {
	str := "<event.source"
	if s.Cap > 0 {
		str += fmt.Sprint(" cap=", s.Cap)
	}
	if len(s.ch) > 0 {
		i := 0
		for _, ch := range s.ch {
			if ch != nil {
				i++
			}
		}
		str += fmt.Sprint(" subs=", i)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (s *Source) Emit(e Event) bool {
	s.RLock()
	defer s.RUnlock()

	result := true
	if e == nil {
		return false
	}
	for _, ch := range s.ch {
		if ch != nil {
			if !s.emit(ch, e) {
				result = false
			}
		}
	}
	return result
}

func (s *Source) Sub() <-chan Event {
	s.Lock()
	defer s.Unlock()
	// s.Cap as zero means unbuffered channels
	ch := make(chan Event, s.Cap)
	s.ch = append(s.ch, ch)
	return ch
}

func (s *Source) Unsub(<-chan Event) {
	s.Lock()
	defer s.Unlock()
	// TODO
}

func (s *Source) Close() error {
	s.Lock()
	defer s.Unlock()
	for _, ch := range s.ch {
		if ch != nil {
			close(ch)
		}
	}
	s.ch = nil
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (s *Source) emit(ch chan<- Event, e Event) bool {
	if cap(ch) > 0 {
		select {
		case ch <- e:
			return true
		default:
			return false
		}
	}
	ch <- e
	return true
}
