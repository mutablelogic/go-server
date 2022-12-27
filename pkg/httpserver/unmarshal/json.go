package unmarshal

import (
	"encoding/json"
	"io"
	"reflect"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC  METHODS

// Decode with a io.Reader which contains JSON and unmarshal into dest. Return the fields
// which were set (for patching)
func WithJson(r io.Reader, dest any) ([]string, error) {
	srcmap := make(map[string]any)
	if err := json.NewDecoder(r).Decode(&srcmap); err != nil {
		return nil, err
	}
	var fields []string
	err := Walk(dest, tag, func(src reflect.Value, name string, extra []string) error {
		// Return if the query is missing the field
		if v, exists := srcmap[name]; !exists {
			return nil
		} else {
			fields = append(fields, name)
			return setValue(src, reflect.ValueOf(v))
		}

	})
	return fields, err
}
