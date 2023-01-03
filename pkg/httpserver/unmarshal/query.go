package unmarshal

import (
	"net/url"
	"reflect"
)

///////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	tag = "decode"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC  METHODS

// Decode with a url query q and unmarshal into dest. Return the fields
// which were set (for patching) from the query
func WithQuery(q url.Values, dest any) ([]string, error) {
	var fields []string
	err := Walk(dest, tag, func(src reflect.Value, name string, extra []string) error {
		// Return if the query is missing the field
		if !q.Has(name) {
			return nil
		} else {
			fields = append(fields, name)
		}
		// Set a source value to a destination value
		if src.Kind() == reflect.Slice {
			return setValue(src, reflect.ValueOf(q[name]))
		} else {
			return setValue(src, reflect.ValueOf(q.Get(name)))
		}
	})
	return fields, err
}
