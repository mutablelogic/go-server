package schema

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	goldap "github.com/go-ldap/ldap/v3"
	// Packages
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Object struct {
	DN         string `json:"dn"`
	url.Values `json:"attrs,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultCapacity = 10
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewObject(v ...string) *Object {
	o := new(Object)
	o.DN = strings.Join(v, ",")
	o.Values = url.Values{}
	return o
}

func NewObjectFromEntry(entry *goldap.Entry) *Object {
	o := NewObject(entry.DN)
	for _, attr := range entry.Attributes {
		o.Values[attr.Name] = attr.Values
	}
	return o
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (o *Object) String() string {
	data, _ := json.MarshalIndent(o, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (o *Object) Set(attr string, values ...string) {
	o.Values[attr] = values
}

// Return gidNumber as an integer, returning -1 if not set
func (o *Object) GroupId() int {
	if v := o.Get("gidNumber"); v != "" {
		if v, err := strconv.ParseInt(v, 10, 32); err == nil && v >= 0 {
			return int(v)
		}
	}
	return -1
}

// Return uidNumber as an integer, returning -1 if not set
func (o *Object) UserId() int {
	if v := o.Get("uidNumber"); v != "" {
		if v, err := strconv.ParseInt(v, 10, 32); err == nil && v >= 0 {
			return int(v)
		}
	}
	return -1
}
