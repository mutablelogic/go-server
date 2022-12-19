package tokenauth

import (
	"fmt"
	"time"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type Token struct {
	Value string    `json:"token,omitempty"`
	Time  time.Time `json:"access_time"`
}

/////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t *Token) String() string {
	str := "<httpserver-token"
	str += fmt.Sprintf(" token=%q", t.Value)
	str += fmt.Sprintf(" access_time=%q", t.Time.Format(time.RFC3339))
	return str + ">"
}
