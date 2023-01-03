package unmarshal

import (
	"reflect"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Walk(v any, tag string, fn func(dest reflect.Value, name string, value []string) error) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return ErrBadParameter.Withf("Walk: requires a pointer to a struct, got %q", rv.Kind())
	} else {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return ErrBadParameter.Withf("Walk: requires a struct, got %q", rv.Kind())
	}
	return walk(rv, tag, nil, fn)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func walk(v reflect.Value, name string, value []string, fn func(dest reflect.Value, name string, value []string) error) error {
	switch v.Kind() {
	case reflect.Struct:
		if v.Type() == typeTime {
			return fn(v, value[0], value[1:])
		}
		for i := 0; i < v.NumField(); i++ {
			value := tagValue(v.Type().Field(i), name)
			if value != nil {
				if err := walk(v.Field(i), tag, value, fn); err != nil {
					return err
				}
			}
		}
	default:
		if len(value) > 0 {
			return fn(v, value[0], value[1:])
		}
	}
	return nil
}
