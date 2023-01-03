package unmarshal

import (
	"reflect"
	"strings"
	"unicode"
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// tagValue returns the value of the field based on tag or field name
// and returns nil if the field should be ignored (not assignable)
func tagValue(field reflect.StructField, tagName string) []string {
	// Check for private field
	if field.Name != "" && unicode.IsLower(rune(field.Name[0])) {
		return nil
	}
	tags := strings.Split(field.Tag.Get(tagName), ",")
	if tags[0] == "-" {
		return nil
	} else if tags[0] == "" {
		return []string{field.Name}
	} else {
		return tags
	}
}
