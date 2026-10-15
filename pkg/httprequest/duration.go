package httprequest

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	typeDuration = reflect.TypeOf(time.Duration(0))

	// hasDurationCache is whether each type has a time.Duration in it
	hasDurationCache sync.Map
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// parseDuration parses a duration, which is integer nanoseconds, or a duration
// string, such as "5s"
func parseDuration(value string) (time.Duration, error) {
	if n, err := strconv.ParseInt(value, 10, 64); err == nil {
		return time.Duration(n), nil
	}
	return time.ParseDuration(value)
}

// hasDuration reports whether the type has a time.Duration in it, as a field
// of a struct, including embedded and nested structs, or as the element of a
// pointer, slice, array or map
func hasDuration(t reflect.Type) bool {
	if v, ok := hasDurationCache.Load(t); ok {
		return v.(bool)
	}
	result := findDuration(t, map[reflect.Type]bool{})
	hasDurationCache.Store(t, result)
	return result
}

func findDuration(t reflect.Type, seen map[reflect.Type]bool) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == typeDuration {
		return true
	} else if seen[t] {
		// A recursive type
		return false
	}
	seen[t] = true
	switch t.Kind() {
	case reflect.Struct:
		for _, field := range reflect.VisibleFields(t) {
			if field.IsExported() && findDuration(field.Type, seen) {
				return true
			}
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		return findDuration(t.Elem(), seen)
	}
	return false
}

// durations returns the JSON value, decoded with json.Number for numbers, with
// the duration strings for the time.Duration values of the type replaced with
// integer nanoseconds, so it can be decoded into the type
func durations(tag string, value any, t reflect.Type) (any, error) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch {
	case t == typeDuration:
		if str, ok := value.(string); ok {
			d, err := parseDuration(str)
			if err != nil {
				return nil, errBadRequest.Withf("%q: Parse error (expected a duration value)", tag)
			}
			return json.Number(strconv.FormatInt(int64(d), 10)), nil
		}
	case t.Kind() == reflect.Struct:
		if m, ok := value.(map[string]any); ok {
			for _, field := range reflect.VisibleFields(t) {
				// The fields of an embedded struct are promoted, so are
				// converted with the struct's own fields
				if !field.IsExported() || (field.Anonymous && jsonTagName(field) == "") {
					continue
				}
				key := jsonName(field)
				if key == "" {
					continue
				} else if item, exists := m[key]; exists {
					converted, err := durations(key, item, field.Type)
					if err != nil {
						return nil, err
					}
					m[key] = converted
				}
			}
		}
	case t.Kind() == reflect.Slice || t.Kind() == reflect.Array:
		if items, ok := value.([]any); ok {
			for i, item := range items {
				converted, err := durations(tag, item, t.Elem())
				if err != nil {
					return nil, err
				}
				items[i] = converted
			}
		}
	case t.Kind() == reflect.Map:
		if m, ok := value.(map[string]any); ok {
			for key, item := range m {
				converted, err := durations(key, item, t.Elem())
				if err != nil {
					return nil, err
				}
				m[key] = converted
			}
		}
	}
	return value, nil
}

// jsonTagName returns the name in the field's JSON tag, or an empty string
func jsonTagName(field reflect.StructField) string {
	name, _, _ := strings.Cut(field.Tag.Get(tagName), ",")
	return name
}
