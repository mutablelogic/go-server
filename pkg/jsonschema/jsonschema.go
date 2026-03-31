package jsonschema

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	// Packages
	upstream "github.com/google/jsonschema-go/jsonschema"
	uuid "github.com/google/uuid"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Schema struct {
	upstream.Schema
	resolved *upstream.Resolved
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var schemaCache sync.Map // reflect.Type -> *Schema

// durationType is the reflect.Type for time.Duration, used to detect duration
// fields and represent them as JSON strings rather than integers.
var durationType = reflect.TypeFor[time.Duration]()

// timeType is the reflect.Type for time.Time, used to detect time fields
// and represent them as JSON strings with format "date-time".
var timeType = reflect.TypeFor[time.Time]()

// uuidType is the reflect.Type for uuid.UUID, used to detect UUID fields
// and represent them as JSON strings with format "uuid".
var uuidType = reflect.TypeFor[uuid.UUID]()

// urlType is the reflect.Type for url.URL, used to detect URL fields
// and represent them as JSON strings with format "uri".
var urlType = reflect.TypeFor[url.URL]()

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Validate validates a JSON value against the schema. The data must be valid JSON.
// It returns nil if validation succeeds, or an error describing the failures.
func (s *Schema) Validate(data json.RawMessage) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	return s.resolved.Validate(v)
}

// Decode unmarshals data into v, applies schema defaults for any missing fields,
// then validates the result. v must be a non-nil pointer; it may point to a
// struct (in which case schema defaults are applied for missing fields), a
// primitive (string, bool, numeric), a slice, or a *time.Duration.
func (s *Schema) Decode(data json.RawMessage, v any) error {
	var instance any
	if err := json.Unmarshal(data, &instance); err != nil {
		return err
	}
	// For objects, use a map[string]any intermediate so ApplyDefaults can fill
	// in defaults and Validate can accept the value (it rejects structs).
	if m, ok := instance.(map[string]any); ok {
		if err := s.resolved.ApplyDefaults(&m); err != nil {
			return err
		}
		if err := s.resolved.Validate(m); err != nil {
			return err
		}
		// Convert duration strings (e.g. "5s") to int64 nanoseconds so that
		// json.Unmarshal can correctly populate time.Duration struct fields.
		if err := convertDurationFields(m, reflect.TypeOf(v)); err != nil {
			return err
		}
		// Re-encode the enriched map and decode into the typed target.
		b, err := json.Marshal(m)
		if err != nil {
			return err
		}
		return json.Unmarshal(b, v)
	}
	// For primitives and arrays, validate directly then decode into target.
	if err := s.resolved.Validate(instance); err != nil {
		return err
	}
	// time.Duration is encoded as a string; parse and assign directly.
	if dp, ok := v.(*time.Duration); ok {
		if str, ok := instance.(string); ok {
			d, err := time.ParseDuration(str)
			if err != nil {
				return fmt.Errorf("invalid duration %q: %w", str, err)
			}
			*dp = d
			return nil
		}
	}
	return json.Unmarshal(data, v)
}

// FromJSON parses a JSON Schema document and returns a resolved *Schema ready
// for use with Validate and Decode. Unlike For[T](), the result is not cached
// since there is no type key to cache against.
func FromJSON(data json.RawMessage) (*Schema, error) {
	var s upstream.Schema
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	res := &Schema{s, nil}
	resolved, err := res.Resolve(nil)
	if err != nil {
		return nil, err
	}
	res.resolved = resolved
	return res, nil
}

// For generates a JSON Schema for the given type T, enriches it with struct
// tag annotations, resolves any $ref references, and caches the result.
// Subsequent calls for the same type return the cached schema at no cost.
func For[T any]() (*Schema, error) {
	t := reflect.TypeFor[T]()
	if v, ok := schemaCache.Load(t); ok {
		return v.(*Schema), nil
	}
	s, err := upstream.For[T](nil)
	if err != nil {
		return nil, err
	}
	// enrichSchema only applies to struct types; skip for primitives, slices, etc.
	ft := t
	for ft.Kind() == reflect.Pointer {
		ft = ft.Elem()
	}
	if ft == durationType {
		// Override the upstream integer schema with a duration string schema.
		s.Type = "string"
		s.Types = nil
		s.Format = "duration"
	} else if ft == timeType {
		// Override the upstream object schema with a date-time string schema.
		s.Type = "string"
		s.Types = nil
		s.Format = "date-time"
		s.Properties = nil
	} else if ft == uuidType {
		// Override the upstream array schema with a uuid string schema.
		s.Type = "string"
		s.Types = nil
		s.Format = "uuid"
		s.Items = nil
		s.MinItems = nil
		s.MaxItems = nil
	} else if ft == urlType {
		// Override the upstream object schema with a uri string schema.
		s.Type = "string"
		s.Types = nil
		s.Format = "uri"
		s.Properties = nil
	} else if ft.Kind() == reflect.Slice && ft.Elem().Kind() == reflect.Uint8 {
		// []byte is represented as a base64 string.
		s.Type = "string"
		s.Types = nil
		s.Format = "byte"
	} else if ft.Kind() == reflect.Struct {
		if err := enrichSchema(s, t); err != nil {
			return nil, err
		}
	}
	res := &Schema{*s, nil}
	resolved, err := res.Resolve(nil)
	if err != nil {
		return nil, err
	}
	res.resolved = resolved

	// Store in cache
	schemaCache.Store(t, res)

	// Return the enriched schema
	return res, nil
}

// MustFor is like [For] but panics if schema generation fails.
func MustFor[T any]() *Schema {
	s, err := For[T]()
	if err != nil {
		panic(fmt.Sprintf("jsonschema.MustFor: %v", err))
	}
	return s
}

// Property returns the schema for the named property, or nil if the property
// is not present. The returned schema wraps the upstream property schema;
// it is not resolved and should only be used for schema generation, not
// for Validate or Decode.
func (s *Schema) Property(name string) *Schema {
	if prop, ok := s.Properties[name]; ok {
		return &Schema{Schema: *prop}
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func enrichSchema(s *upstream.Schema, t reflect.Type) error {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip if the field is unexported or has a json tag that excludes it.
		jsonName := jsonFieldName(field)
		if jsonName == "-" || jsonName == "" {
			continue
		}

		// Look up the property in the schema. If it's not present, skip it.
		prop := s.Properties[jsonName]
		if prop == nil {
			continue
		}

		// Dereference pointer types once for type-based tag dispatch.
		ft := field.Type
		for ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}

		// time.Duration fields are represented as duration strings, not integers.
		if ft == durationType {
			prop.Type = "string"
			prop.Types = nil
			prop.Format = "duration"
		}

		// time.Time fields are represented as date-time strings, not objects.
		if ft == timeType {
			prop.Type = "string"
			prop.Types = nil
			prop.Format = "date-time"
			prop.Properties = nil
		}

		// uuid.UUID fields are represented as uuid strings, not byte arrays.
		if ft == uuidType {
			prop.Type = "string"
			prop.Types = nil
			prop.Format = "uuid"
			prop.Items = nil
			prop.MinItems = nil
			prop.MaxItems = nil
		}

		// url.URL fields are represented as uri strings, not objects.
		if ft == urlType {
			prop.Type = "string"
			prop.Types = nil
			prop.Format = "uri"
			prop.Properties = nil
		}

		// []byte fields are represented as base64 strings.
		if ft.Kind() == reflect.Slice && ft.Elem().Kind() == reflect.Uint8 {
			prop.Type = "string"
			prop.Types = nil
			prop.Format = "byte"
		}

		if vals := parseEnumTag(field.Tag.Get("enum")); len(vals) > 0 {
			prop.Enum = vals
		}
		if v := field.Tag.Get("format"); v != "" {
			prop.Format = v
		}

		// Parse the "default" tag according to the field's type, and set the schema default if successful.
		if v := field.Tag.Get("default"); v != "" {
			if raw := marshalDefault(field.Type, v); raw != nil {
				prop.Default = raw
			}
		}

		// Default to the "help" tag if Description is not already set.
		if prop.Description == "" {
			prop.Description = field.Tag.Get("help")
		}

		// required / optional via kong standalone tags.
		// A field with a default is implicitly optional (ApplyDefaults only fills
		// in non-required fields), unless explicitly tagged required:"".
		_, isExplicitRequired := field.Tag.Lookup("required")
		_, isExplicitOptional := field.Tag.Lookup("optional")
		hasDefault := prop.Default != nil
		if isExplicitOptional || (hasDefault && !isExplicitRequired) {
			s.Required = removeString(s.Required, jsonName)
		} else if isExplicitRequired {
			s.Required = appendUnique(s.Required, jsonName)
		}

		// A pointer field that is required (explicitly or by upstream default)
		// must not allow null: remove "null" from Types.
		if field.Type.Kind() == reflect.Pointer {
			isRequired := false
			for _, r := range s.Required {
				if r == jsonName {
					isRequired = true
					break
				}
			}
			if isRequired {
				prop.Types = removeString(prop.Types, "null")
				if prop.Type == "null" {
					prop.Type = ""
				}
			}
		}

		// Slice fields should not allow null. The upstream library adds "null"
		// because Go slices have a nil zero value, but in JSON APIs an absent
		// slice is omitted or [], never null.
		if field.Type.Kind() == reflect.Slice {
			prop.Types = removeString(prop.Types, "null")
			if prop.Type == "null" {
				prop.Type = ""
			}
		}

		if v := field.Tag.Get("min"); v != "" {
			applyMin(prop, ft, v)
		}
		if v := field.Tag.Get("max"); v != "" {
			applyMax(prop, ft, v)
		}
		if _, ok := field.Tag.Lookup("deprecated"); ok {
			prop.Deprecated = true
		}
		if v := field.Tag.Get("pattern"); v != "" {
			prop.Pattern = v
		}
		if _, ok := field.Tag.Lookup("readonly"); ok {
			prop.ReadOnly = true
		}
		if v := field.Tag.Get("example"); v != "" {
			var parsed any
			if err := json.Unmarshal([]byte(v), &parsed); err == nil {
				prop.Examples = append(prop.Examples, parsed)
			} else if raw := marshalDefault(field.Type, v); raw != nil {
				var decoded any
				if err := json.Unmarshal(raw, &decoded); err == nil {
					prop.Examples = append(prop.Examples, decoded)
				}
			}
		}

		// If the field's type is a struct or pointer to struct, recursively enrich it.
		if ft.Kind() == reflect.Struct {
			if err := enrichSchema(prop, ft); err != nil {
				return err
			}
		}
	}

	// Return success
	return nil
}

// marshalDefault converts a string tag value to json.RawMessage using the
// field type to choose the correct JSON representation.
func marshalDefault(t reflect.Type, val string) json.RawMessage {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	// time.Duration defaults are stored as duration strings (e.g. "5s"), not integers.
	if t == durationType {
		if b, err := json.Marshal(val); err == nil {
			return b
		}
		return nil
	}
	switch t.Kind() {
	case reflect.Bool:
		if b, err := strconv.ParseBool(val); err == nil {
			if b {
				return json.RawMessage("true")
			}
			return json.RawMessage("false")
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if n, err := strconv.ParseInt(val, 10, 64); err == nil {
			return json.RawMessage(strconv.FormatInt(n, 10))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if n, err := strconv.ParseUint(val, 10, 64); err == nil {
			return json.RawMessage(strconv.FormatUint(n, 10))
		}
	case reflect.Float32, reflect.Float64:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			if b, err := json.Marshal(f); err == nil {
				return b
			}
		}
	}
	// fallback: encode as JSON string
	if b, err := json.Marshal(val); err == nil {
		return b
	}
	return nil
}

func applyMin(prop *upstream.Schema, ft reflect.Type, v string) {
	switch ft.Kind() {
	case reflect.String:
		if n, err := strconv.Atoi(v); err == nil {
			prop.MinLength = &n
		}
	case reflect.Slice, reflect.Array:
		if n, err := strconv.Atoi(v); err == nil {
			prop.MinItems = &n
		}
	default:
		if isNumericKind(ft.Kind()) {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				prop.Minimum = &f
			}
		}
	}
}

func applyMax(prop *upstream.Schema, ft reflect.Type, v string) {
	switch ft.Kind() {
	case reflect.String:
		if n, err := strconv.Atoi(v); err == nil {
			prop.MaxLength = &n
		}
	case reflect.Slice, reflect.Array:
		if n, err := strconv.Atoi(v); err == nil {
			prop.MaxItems = &n
		}
	default:
		if isNumericKind(ft.Kind()) {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				prop.Maximum = &f
			}
		}
	}
}

func isNumericKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

func appendUnique(ss []string, s string) []string {
	for _, v := range ss {
		if v == s {
			return ss
		}
	}
	return append(ss, s)
}

func removeString(ss []string, s string) []string {
	out := ss[:0:0] // fresh slice, no shared backing array
	for _, v := range ss {
		if v != s {
			out = append(out, v)
		}
	}
	return out
}

// convertDurationFields walks the struct type of v and converts any time.Duration
// field values in m from duration strings (e.g. "5s") to int64 nanoseconds, so
// that json.Unmarshal can correctly populate time.Duration struct fields.
// It recurses into nested struct fields whose map values are map[string]any.
func convertDurationFields(m map[string]any, t reflect.Type) error {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		ft := field.Type
		for ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}
		key := jsonFieldName(field)
		if key == "" || key == "-" {
			continue
		}
		val, ok := m[key]
		if !ok {
			continue
		}
		if ft == durationType {
			if str, ok := val.(string); ok {
				d, err := time.ParseDuration(str)
				if err != nil {
					return fmt.Errorf("field %q: invalid duration %q: %w", key, str, err)
				}
				m[key] = int64(d)
			}
		} else if ft.Kind() == reflect.Struct {
			if nested, ok := val.(map[string]any); ok {
				if err := convertDurationFields(nested, ft); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// parseEnumTag splits a comma-separated enum tag value into a slice of
// non-empty, trimmed strings. Returns nil or an empty slice if the tag
// contains no non-whitespace values.
func parseEnumTag(tag string) []any {
	if tag == "" {
		return nil
	}
	parts := strings.Split(tag, ",")
	vals := make([]any, 0, len(parts))
	for _, v := range parts {
		if v = strings.TrimSpace(v); v != "" {
			vals = append(vals, v)
		}
	}
	return vals
}

// jsonFieldName extracts the JSON field name from a struct field's "json" tag.
func jsonFieldName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "" {
		return f.Name
	}
	name, _, _ := strings.Cut(tag, ",")
	if name == "" {
		return f.Name
	}
	return name
}
