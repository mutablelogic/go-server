package schema

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Attribute describes a single configuration field within a [Schema].
type Attribute struct {
	// Name is the unique name of the field within the resource schema.
	Name string `json:"name"`

	// Type is the value type (e.g. "string", "int", "bool", "duration",
	// "[]string").
	Type string `json:"type"`

	// Description is a human-readable explanation of the field.
	Description string `json:"description"`

	// Required indicates the field must be set by the caller.
	Required bool `json:"required"`

	// Default is the value used when the caller does not set the field.
	// It must be assignable to Type.
	Default any `json:"default,omitempty"`

	// Sensitive marks the field as containing secrets that should not
	// appear in logs or plan output.
	Sensitive bool `json:"sensitive"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Attributes uses reflection to convert a struct into a slice of [Attribute],
// reading struct tags to determine each field's properties:
//
//   - name:"<name>" — attribute name (required; omitted or name:"-" skips the field)
//   - help:"<text>" — human-readable description
//   - default:"<value>" — default value (field is optional when present)
//   - required:"" — marks the attribute as required (ignored if default is set)
//   - type:"<type>" — overrides the inferred type string (e.g. "file", "url")
//   - sensitive:"" — marks the field as containing secrets; values are redacted
//     in plan output, logs and state serialisation
//   - embed:"" with prefix:"<prefix>" — flatten nested struct, prepending prefix
//     to all child attribute names
func Attributes(resource any) []Attribute {
	rv := reflect.ValueOf(resource)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}
	return structAttributes(rv.Type(), "")
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func structAttributes(t reflect.Type, prefix string) []Attribute {
	var attrs []Attribute

	for i := range t.NumField() {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Handle structs with embed:"" tag — flatten with prefix
		if field.Type.Kind() == reflect.Struct && hasTag(field.Tag, "embed") {
			childPrefix := prefix + field.Tag.Get("prefix")
			attrs = append(attrs, structAttributes(field.Type, childPrefix)...)
			continue
		}

		// Skip fields tagged name:"-" or with no name tag
		name, hasName := field.Tag.Lookup("name")
		if !hasName || name == "-" {
			continue
		}

		// Prepend the prefix to the attribute name
		name = prefix + name

		// Determine the type string
		typeName := field.Tag.Get("type")
		if typeName == "" {
			typeName = goTypeString(field.Type)
		}

		// Build the attribute
		def, hasDef := field.Tag.Lookup("default")
		attr := Attribute{
			Name:        name,
			Type:        typeName,
			Description: field.Tag.Get("help"),
			Required:    !hasDef && hasTag(field.Tag, "required"),
			Sensitive:   hasTag(field.Tag, "sensitive"),
		}
		if hasDef {
			attr.Default = def
		}

		attrs = append(attrs, attr)
	}
	return attrs
}

// goTypeString returns a human-readable type name for a reflect.Type.
func goTypeString(t reflect.Type) string {
	// Check for well-known types
	switch t {
	case reflect.TypeOf(time.Duration(0)):
		return "duration"
	case reflect.TypeOf((*time.Time)(nil)).Elem():
		return "time"
	}

	// Unwrap pointer
	if t.Kind() == reflect.Ptr {
		return goTypeString(t.Elem())
	}

	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "uint"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.Slice:
		return "[]" + goTypeString(t.Elem())
	case reflect.Interface:
		return t.String()
	default:
		return t.String()
	}
}

// tagDefault returns the default value from a struct tag, or nil if empty.
func tagDefault(v string) any {
	if v == "" {
		return nil
	}
	return v
}

// hasTag reports whether the struct tag contains the named key,
// even if its value is empty (e.g. `embed:""`).
func hasTag(tag reflect.StructTag, key string) bool {
	_, ok := tag.Lookup(key)
	return ok
}

///////////////////////////////////////////////////////////////////////////////
// STATE

// StateOf uses reflection to extract the current field values from a struct
// and return them as a [State] map keyed by attribute name. It follows the
// same struct-tag rules as [Attributes]: fields without a name tag or with
// name:"-" are skipped, and embedded structs are flattened with their prefix.
//
// Interface fields that implement [ResourceInstance] (or any interface with
// a Name() string method) are stored by name. Other interface fields are
// skipped.
//
// Duration values are stored as their string representation (e.g. "5m0s").
func StateOf(resource any) State {
	rv := reflect.ValueOf(resource)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}
	s := make(State)
	structState(rv, "", s)
	return s
}

func structState(rv reflect.Value, prefix string, s State) {
	t := rv.Type()
	for i := range t.NumField() {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Handle structs with embed:"" tag — flatten with prefix
		if field.Type.Kind() == reflect.Struct && hasTag(field.Tag, "embed") {
			childPrefix := prefix + field.Tag.Get("prefix")
			structState(rv.Field(i), childPrefix, s)
			continue
		}

		// Skip fields tagged name:"-" or with no name tag
		name, hasName := field.Tag.Lookup("name")
		if !hasName || name == "-" {
			continue
		}

		// Interface fields: if the value implements ResourceInstance, store
		// the instance name as a reference; otherwise skip (not serialisable)
		if field.Type.Kind() == reflect.Interface {
			v := rv.Field(i)
			if v.IsNil() {
				continue
			}
			if ri, ok := v.Interface().(ResourceInstance); ok {
				name = prefix + name
				s[name] = ri.Name()
			}
			continue
		}

		// Prepend the prefix to the key
		name = prefix + name

		// Extract the value, handling pointers and well-known types
		val := rv.Field(i)
		s[name] = stateValue(val)
	}
}

// stateValue converts a reflect.Value to a plain Go value suitable for
// storage in a [State] map.
func stateValue(v reflect.Value) any {
	// Dereference pointer
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	// Duration → string
	if v.Type() == reflect.TypeOf(time.Duration(0)) {
		return v.Interface().(time.Duration).String()
	}

	return v.Interface()
}

///////////////////////////////////////////////////////////////////////////////
// VALIDATE REFERENCES

// ValidateRefs walks a resource config struct and validates every
// [ResourceInstance] reference field against its struct tags:
//
//   - required:"" — the reference must be non-nil
//   - type:"<name>" — the referenced instance's [Resource.Name] must match
//
// It returns a combined error for all failing fields, or nil if valid.
func ValidateRefs(resource any) error {
	rv := reflect.ValueOf(resource)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}
	return validateRefs(rv, "")
}

///////////////////////////////////////////////////////////////////////////////
// REFERENCES

// ReferencesOf walks a resource config struct and returns the [ResourceInstance.Name]
// of every non-nil [ResourceInstance] reference field. The result is the set
// of dependency names that must be applied before this resource.
func ReferencesOf(resource any) []string {
	rv := reflect.ValueOf(resource)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}
	var refs []string
	referencesOf(rv, &refs)
	if len(refs) == 0 {
		return nil
	}
	return refs
}

func referencesOf(rv reflect.Value, refs *[]string) {
	t := rv.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		// Recurse into embedded structs
		if field.Type.Kind() == reflect.Struct && hasTag(field.Tag, "embed") {
			referencesOf(rv.Field(i), refs)
			continue
		}
		// Only check interface fields with a name
		if field.Type.Kind() != reflect.Interface {
			continue
		}
		name, hasName := field.Tag.Lookup("name")
		if !hasName || name == "-" {
			continue
		}
		v := rv.Field(i)
		if v.IsNil() {
			continue
		}
		if ri, ok := v.Interface().(ResourceInstance); ok {
			*refs = append(*refs, ri.Name())
		}
	}
}

func validateRefs(rv reflect.Value, prefix string) error {
	t := rv.Type()
	var errs []error

	for i := range t.NumField() {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Handle structs with embed:"" tag — recurse with prefix
		if field.Type.Kind() == reflect.Struct && hasTag(field.Tag, "embed") {
			childPrefix := prefix + field.Tag.Get("prefix")
			if err := validateRefs(rv.Field(i), childPrefix); err != nil {
				errs = append(errs, err)
			}
			continue
		}

		// Only check interface fields
		if field.Type.Kind() != reflect.Interface {
			continue
		}

		// Skip fields tagged name:"-" or with no name tag
		name, hasName := field.Tag.Lookup("name")
		if !hasName || name == "-" {
			continue
		}
		name = prefix + name

		v := rv.Field(i)

		// Check required
		if v.IsNil() {
			if hasTag(field.Tag, "required") {
				errs = append(errs, fmt.Errorf("%s: required", name))
			}
			continue
		}

		// Check type constraint against Resource().Name()
		wantType := field.Tag.Get("type")
		if wantType == "" {
			continue
		}
		if ri, ok := v.Interface().(ResourceInstance); ok {
			if gotType := ri.Resource().Name(); gotType != wantType {
				errs = append(errs, fmt.Errorf("%s: must be of type %q, got %q", name, wantType, gotType))
			}
		}
	}

	return errors.Join(errs...)
}
