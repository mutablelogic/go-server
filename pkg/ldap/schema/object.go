package schema

import (
	"encoding/json"
	"net/url"
	"strings"

	// Packages
	pg "github.com/djthorpe/go-pg"
	ldap "github.com/go-ldap/ldap/v3"
)

//////////////////////////////////////////////////////////////////////////////////
// TYPES

type Object struct {
	DN         string `json:"dn"`
	url.Values `json:"attrs,omitempty"`
}

type ObjectListRequest struct {
	pg.OffsetLimit
	Filter *string `json:"filter,omitempty"`
}

type ObjectList struct {
	Count uint64    `json:"count"`
	Body  []*Object `json:"body,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewObject(v ...string) *Object {
	o := new(Object)
	o.DN = strings.Join(v, ",")
	o.Values = url.Values{}
	return o
}

func NewObjectFromEntry(entry *ldap.Entry) *Object {
	o := NewObject(entry.DN)
	for _, attr := range entry.Attributes {
		o.Values[attr.Name] = attr.Values
	}
	return o
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (o *Object) String() string {
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (o *ObjectList) String() string {
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (o *ObjectListRequest) String() string {
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
