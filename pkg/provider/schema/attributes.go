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
	Description string `json:"description,omitempty"`

	// Required indicates the field must be set by the caller.
	Required bool `json:"required,omitempty"`

	// Default is the value used when the caller does not set the field.
	// It must be assignable to Type.
	Default any `json:"default,omitempty"`

	// Sensitive marks the field as containing secrets that should not
	// appear in logs or plan output.
	Sensitive bool `json:"sensitive,omitempty"`

	// ReadOnly marks the field as computed by the provider. It is not
	// settable by the caller and is populated during [Apply].
	ReadOnly bool `json:"readonly,omitempty"`

	// Reference indicates this attribute is a dependency on another
	// resource instance, resolved by name at decode time.
	Reference bool `json:"reference,omitempty"`
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
func AttributesOf(resource any) []Attribute {
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
			ReadOnly:    hasTag(field.Tag, "readonly"),
			Sensitive:   hasTag(field.Tag, "sensitive"),
			Reference:   field.Type.Kind() == reflect.Interface || (field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Interface),
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
	case durationType:
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
	case reflect.Map:
		return "map[" + goTypeString(t.Key()) + "]" + goTypeString(t.Elem())
	case reflect.Interface:
		return "ref"
	default:
		return t.String()
	}
}

// hasTag reports whether the struct tag contains the named key,
// even if its value is empty (e.g. `embed:""`).
func hasTag(tag reflect.StructTag, key string) bool {
	_, ok := tag.Lookup(key)
	return ok
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

		name, hasName := field.Tag.Lookup("name")
		if !hasName || name == "-" {
			continue
		}

		// Single interface field — one reference
		if field.Type.Kind() == reflect.Interface {
			v := rv.Field(i)
			if v.IsNil() {
				continue
			}
			if ri, ok := v.Interface().(ResourceInstance); ok {
				*refs = append(*refs, ri.Name())
			}
			continue
		}

		// Slice-of-interface field — multiple references
		if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Interface {
			slice := rv.Field(i)
			for j := range slice.Len() {
				if ri, ok := slice.Index(j).Interface().(ResourceInstance); ok {
					*refs = append(*refs, ri.Name())
				}
			}
		}
	}
}

// ReferencesFromState returns the instance names stored as references in the
// state, by matching reference attributes from the schema against state keys.
// Unlike [ReferencesOf], this works from serialised state without needing the
// live struct with populated interface fields.
func ReferencesFromState(attrs []Attribute, state State) []string {
	var refs []string
	for _, attr := range attrs {
		if !attr.Reference {
			continue
		}
		switch v := state[attr.Name].(type) {
		case string:
			if v != "" {
				refs = append(refs, v)
			}
		case []string:
			refs = append(refs, v...)
		case []any:
			for _, elem := range v {
				if s, ok := elem.(string); ok && s != "" {
					refs = append(refs, s)
				}
			}
		}
	}
	if len(refs) == 0 {
		return nil
	}
	return refs
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

		// Skip fields tagged name:"-" or with no name tag
		name, hasName := field.Tag.Lookup("name")
		if !hasName || name == "-" {
			continue
		}
		name = prefix + name

		// Single interface field
		if field.Type.Kind() == reflect.Interface {
			v := rv.Field(i)
			if v.IsNil() {
				if hasTag(field.Tag, "required") {
					errs = append(errs, fmt.Errorf("%s: required", name))
				}
				continue
			}
			wantType := field.Tag.Get("type")
			if wantType == "" {
				continue
			}
			if ri, ok := v.Interface().(ResourceInstance); ok {
				if gotType := ri.Resource().Name(); gotType != wantType {
					errs = append(errs, fmt.Errorf("%s: must be of type %q, got %q", name, wantType, gotType))
				}
			}
			continue
		}

		// Slice-of-interface field
		if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Interface {
			slice := rv.Field(i)
			if slice.Len() == 0 {
				if hasTag(field.Tag, "required") {
					errs = append(errs, fmt.Errorf("%s: required", name))
				}
				continue
			}
			wantType := field.Tag.Get("type")
			if wantType == "" {
				continue
			}
			for j := range slice.Len() {
				if ri, ok := slice.Index(j).Interface().(ResourceInstance); ok {
					if gotType := ri.Resource().Name(); gotType != wantType {
						errs = append(errs, fmt.Errorf("%s[%d]: must be of type %q, got %q", name, j, wantType, gotType))
					}
				}
			}
		}
	}

	return errors.Join(errs...)
}

// ValidateRequired checks that every field tagged required:"" (without a
// default:"" tag) has a non-zero value.  Reference (interface) fields are
// skipped — those are handled by [ValidateRefs].
func ValidateRequired(resource any) error {
	rv := reflect.ValueOf(resource)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}
	return validateRequired(rv, "")
}

func validateRequired(rv reflect.Value, prefix string) error {
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
			if err := validateRequired(rv.Field(i), childPrefix); err != nil {
				errs = append(errs, err)
			}
			continue
		}

		// Skip interface fields (handled by ValidateRefs)
		if field.Type.Kind() == reflect.Interface {
			continue
		}
		if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Interface {
			continue
		}

		// Skip fields tagged name:"-" or with no name tag
		name, hasName := field.Tag.Lookup("name")
		if !hasName || name == "-" {
			continue
		}
		name = prefix + name

		// Skip readonly fields (set by the provider, not the caller)
		if hasTag(field.Tag, "readonly") {
			continue
		}

		// Skip fields with a default (they always have a value)
		if _, hasDef := field.Tag.Lookup("default"); hasDef {
			continue
		}

		// Check required
		if hasTag(field.Tag, "required") && rv.Field(i).IsZero() {
			errs = append(errs, fmt.Errorf("%s: required", name))
		}
	}

	return errors.Join(errs...)
}
