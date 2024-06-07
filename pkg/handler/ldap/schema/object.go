package schema

import (
	"encoding/json"

	// Packages
	goldap "github.com/go-ldap/ldap/v3"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Object struct {
	*goldap.Entry
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (o *Object) MarshalJSON() ([]byte, error) {
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

func (o *Object) String() string {
	data, _ := json.MarshalIndent(o, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (o *Object) Set(name string, values ...string) {
	o.Entry.Attributes = []*goldap.EntryAttribute{
		&goldap.EntryAttribute{Name: name, Values: values},
	}
}
