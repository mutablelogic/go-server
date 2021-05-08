package location

import "fmt"

// Location stores current location
type Location struct {
	Name string
}

// NewLocation returns a location object
func NewLocation(name string) *Location {
	l := new(Location)
	l.Name = name
	return l
}

// String returns the current location
func (l *Location) String() string {
	return fmt.Sprint(l.Name)
}
