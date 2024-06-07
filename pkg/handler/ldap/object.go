package ldap

import (
	// Packages
	"encoding/json"

	goldap "github.com/go-ldap/ldap/v3"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type object struct {
	*goldap.Entry
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newObject(entry *goldap.Entry) *object {
	return &object{Entry: entry}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (o *object) MarshalJSON() ([]byte, error) {
	j := struct {
		DN    string              `json:"dn"`
		Attrs map[string][]string `json:"attrs"`
	}{
		DN:    o.Entry.DN,
		Attrs: make(map[string][]string),
	}
	for _, attr := range o.Entry.Attributes {
		j.Attrs[attr.Name] = attr.Values
	}
	return json.Marshal(j)
}

func (o *object) String() string {
	data, _ := json.MarshalIndent(o, "", "  ")
	return string(data)
}
