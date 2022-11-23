package event

import (
	"fmt"
	"sync"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Source of events, which can be subscribed to, or unsubscribed from
type Source struct {
	Cap  int          // Capacity of any created channels, setting to zero means unbuffered
	ch   []chan Event // subscribed channels
	lock sync.RWMutex
}

// Compile time check
var _ EventSource = (*Source)(nil)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// Return source as a string object
func (s *Source) String() string {
	str := "<event.source"
	if s.Cap > 0 {
		str += fmt.Sprint(" cap=", s.Cap)
	}
	if l := s.Len(); l > 0 {
		str += fmt.Sprint(" len=", l)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Emit the event to subscribers. Returns false if any channel failed, due to
// buffered channel being full or some other channel issue.
func (s *Source) Emit(e Event) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	result := true
	if e == nil {
		return false
	}
	for _, ch := range s.ch {
		if ch != nil {
			if !e.Emit(ch) {
				result = false
			}
		}
	}
	return result
}

// Subscribe to the source of events. Returns the channel.
func (s *Source) Sub() <-chan Event {
	s.lock.Lock()
	defer s.lock.Unlock()
	// s.Cap as zero means unbuffered channels
	ch := make(chan Event, s.Cap)
	s.ch = append(s.ch, ch)
	return ch
}

// Unsubscribe a channel and close it. Removes the channel from the array
// of channels.
func (s *Source) Unsub(ch <-chan Event) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Identify channel
	j := -1
	for i := range s.ch {
		if s.ch[i] == ch && ch != nil {
			j = i
			break
		}
	}

	// Check to make sure channel was closed
	if j < 0 {
		panic("Unsub called with invalid channel")
	}

	// Close channel
	close(s.ch[j])

	// Remove channel from slice
	s.ch[j] = s.ch[len(s.ch)-1] // Copy last element to index j
	s.ch[len(s.ch)-1] = nil     // Erase last element
	s.ch = s.ch[:len(s.ch)-1]   // Truncate slice
}

// Return number of subscribers
func (s *Source) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.ch)
}

// Close all subscribed channels. Returns any errors (usually nil)
func (s *Source) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, ch := range s.ch {
		if ch != nil {
			close(ch)
		}
	}
	s.ch = nil
	return nil
}
