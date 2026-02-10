package schema

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// State is an opaque snapshot of a resource after it has been applied.
type State map[string]any

// Resolver maps a resource-instance name (as stored by [StateOf]) to the
// live [ResourceInstance], or returns nil if it does not exist.
type Resolver func(name string) ResourceInstance

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	durationType = reflect.TypeOf(time.Duration(0))
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Decode sets the fields of the struct pointed to by v from the state map,
// using the same struct-tag conventions as [StateOf] and [Attributes].
// v must be a pointer to a struct. Fields whose keys are absent from the
// state are left untouched.
//
// Interface fields that hold resource references are resolved using the
// [Resolver]. When the resolver is nil, any interface field whose key is
// present in the state will cause an error. When a resolver returns nil
// for a field marked required:"", an error is returned.
func (s State) Decode(v any, resolve Resolver) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("Decode: expected pointer to struct, got %T", v)
	}
	return decodeStruct(s, rv.Elem(), "", resolve)
}

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

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

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

		// Slice-of-interface fields: store []string of instance names
		if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Interface {
			slice := rv.Field(i)
			if slice.IsNil() || slice.Len() == 0 {
				continue
			}
			names := make([]string, 0, slice.Len())
			for j := range slice.Len() {
				if ri, ok := slice.Index(j).Interface().(ResourceInstance); ok {
					names = append(names, ri.Name())
				}
			}
			s[prefix+name] = names
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
	if v.Type() == durationType {
		return v.Interface().(time.Duration).String()
	}

	return v.Interface()
}

func decodeStruct(s State, rv reflect.Value, prefix string, resolve Resolver) error {
	t := rv.Type()
	for i := range t.NumField() {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Handle structs with embed:"" tag — recurse with prefix
		if field.Type.Kind() == reflect.Struct && hasTag(field.Tag, "embed") {
			childPrefix := prefix + field.Tag.Get("prefix")
			if err := decodeStruct(s, rv.Field(i), childPrefix, resolve); err != nil {
				return err
			}
			continue
		}

		// Skip fields tagged name:"-" or with no name tag
		name, hasName := field.Tag.Lookup("name")
		if !hasName || name == "-" {
			continue
		}
		name = prefix + name

		// Skip readonly fields — they are set by the provider, not the caller
		if hasTag(field.Tag, "readonly") {
			continue
		}

		// Interface fields — resolve via the Resolver callback
		if field.Type.Kind() == reflect.Interface {
			refName, _ := s[name].(string)
			if refName == "" {
				if hasTag(field.Tag, "required") {
					return fmt.Errorf("Decode: field %q: required reference not set", name)
				}
				continue
			}
			if resolve == nil {
				return fmt.Errorf("Decode: field %q: no resolver for reference %q", name, refName)
			}
			inst := resolve(refName)
			if inst == nil {
				if hasTag(field.Tag, "required") {
					return fmt.Errorf("Decode: field %q: reference %q not found", name, refName)
				}
				continue
			}
			rv.Field(i).Set(reflect.ValueOf(inst))
			continue
		}

		// Slice-of-interface fields — resolve each element via Resolver
		if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Interface {
			raw, exists := s[name]
			if !exists {
				continue
			}
			var names []string
			switch v := raw.(type) {
			case []string:
				names = v
			case []any:
				for _, elem := range v {
					if str, ok := elem.(string); ok {
						names = append(names, str)
					}
				}
			default:
				return fmt.Errorf("Decode: field %q: expected []string, got %T", name, raw)
			}
			if resolve == nil && len(names) > 0 {
				return fmt.Errorf("Decode: field %q: no resolver for references", name)
			}
			result := reflect.MakeSlice(field.Type, 0, len(names))
			for _, refName := range names {
				inst := resolve(refName)
				if inst == nil {
					return fmt.Errorf("Decode: field %q: reference %q not found", name, refName)
				}
				result = reflect.Append(result, reflect.ValueOf(inst))
			}
			rv.Field(i).Set(result)
			continue
		}

		// Look up the key in the state map; when absent, fall back to
		// the struct tag default value (if any).
		val, exists := s[name]
		if !exists {
			if def, hasDef := field.Tag.Lookup("default"); hasDef {
				if err := setField(rv.Field(i), def); err != nil {
					return fmt.Errorf("Decode: field %q: default %q: %w", name, def, err)
				}
			}
			continue
		}

		// Set the field value
		if err := setField(rv.Field(i), val); err != nil {
			return fmt.Errorf("Decode: field %q: %w", name, err)
		}
	}
	return nil
}

// setField converts val to the type of dst and sets it.
func setField(dst reflect.Value, val any) error {
	if val == nil {
		return nil
	}

	// Dereference pointer field — allocate if needed
	if dst.Kind() == reflect.Ptr {
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		dst = dst.Elem()
	}

	// Duration: stored as string (e.g. "5m0s")
	if dst.Type() == durationType {
		s, ok := val.(string)
		if !ok {
			return fmt.Errorf("expected string for duration, got %T", val)
		}
		d, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		dst.Set(reflect.ValueOf(d))
		return nil
	}

	src := reflect.ValueOf(val)

	// If directly assignable, just set it
	if src.Type().AssignableTo(dst.Type()) {
		dst.Set(src)
		return nil
	}

	// If convertible (e.g. float64 → int from JSON numbers), convert
	if src.Type().ConvertibleTo(dst.Type()) {
		dst.Set(src.Convert(dst.Type()))
		return nil
	}

	// String → bool
	if dst.Kind() == reflect.Bool {
		if s, ok := val.(string); ok {
			b, err := strconv.ParseBool(s)
			if err != nil {
				return err
			}
			dst.SetBool(b)
			return nil
		}
	}

	// String → numeric types
	if s, ok := val.(string); ok {
		switch dst.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n, err := strconv.ParseInt(s, 10, dst.Type().Bits())
			if err != nil {
				return err
			}
			dst.SetInt(n)
			return nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n, err := strconv.ParseUint(s, 10, dst.Type().Bits())
			if err != nil {
				return err
			}
			dst.SetUint(n)
			return nil
		case reflect.Float32, reflect.Float64:
			f, err := strconv.ParseFloat(s, dst.Type().Bits())
			if err != nil {
				return err
			}
			dst.SetFloat(f)
			return nil
		}
	}

	// []any → []T (element-wise conversion for JSON round-trip)
	if dst.Kind() == reflect.Slice && src.Kind() == reflect.Slice {
		result := reflect.MakeSlice(dst.Type(), src.Len(), src.Len())
		for i := range src.Len() {
			if err := setField(result.Index(i), src.Index(i).Interface()); err != nil {
				return fmt.Errorf("index %d: %w", i, err)
			}
		}
		dst.Set(result)
		return nil
	}

	// map[string]any → map[string]T (element-wise conversion for JSON round-trip)
	if dst.Kind() == reflect.Map && src.Kind() == reflect.Map && dst.Type().Key().Kind() == reflect.String {
		result := reflect.MakeMap(dst.Type())
		iter := src.MapRange()
		for iter.Next() {
			elemPtr := reflect.New(dst.Type().Elem())
			if err := setField(elemPtr.Elem(), iter.Value().Interface()); err != nil {
				return fmt.Errorf("key %v: %w", iter.Key().Interface(), err)
			}
			result.SetMapIndex(iter.Key(), elemPtr.Elem())
		}
		dst.Set(result)
		return nil
	}

	return fmt.Errorf("cannot assign %T to %s", val, dst.Type())
}
