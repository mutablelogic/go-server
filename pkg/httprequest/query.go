package httprequest

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	// Packages
	"github.com/mutablelogic/go-server/pkg/httpresponse"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	tagName       = "json"
	errBadRequest = httpresponse.ErrBadRequest
)

var (
	typeTime = reflect.TypeOf(time.Time{})
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Read the request query parameters into a structure
func Query(q url.Values, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return errBadRequest.With("v must be a pointer")
	} else {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return errBadRequest.With("v must be a pointer to a struct")
	}

	// Enumerate fields
	fields := reflect.VisibleFields(rv.Type())
	if len(fields) == 0 {
		return errBadRequest.With("v has no public fields")
	}
	for _, field := range fields {
		// TODO: support nested structs better. For example, if an embedded struct is called Format
		// and has a field called Format, it will be ignored. Maybe this is a golang bug...
		// need to check.
		tag := jsonName(field)
		if tag == "" {
			continue
		}
		v := rv.FieldByName(field.Name)
		if !v.CanSet() {
			continue
		}
		if value, exists := q[tag]; exists {
			if err := setQueryValue(tag, rv.FieldByIndex(field.Index), value); err != nil {
				return err
			}
		} else {
			if err := setQueryValue(tag, rv.FieldByIndex(field.Index), nil); err != nil {
				return err
			}
		}
	}

	// Return success
	return nil
}

func jsonName(field reflect.StructField) string {
	tag := field.Tag.Get(tagName)
	if tag == "-" {
		return ""
	}
	if fields := strings.Split(tag, ","); len(fields) > 0 && fields[0] != "" {
		return fields[0]
	}
	return field.Name
}

func setQueryValue(tag string, v reflect.Value, value []string) error {
	// Set zero-value
	if len(value) == 0 {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	// Create a new value for pointers
	if v.Kind() == reflect.Ptr {
		v.Set(reflect.New(v.Type().Elem()))
		v = v.Elem()
	}
	// Set the value
	switch v.Kind() {
	case reflect.String:
		if len(value) > 0 {
			v.SetString(value[0])
		}
	case reflect.Bool:
		value, err := strconv.ParseBool(value[0])
		if err != nil {
			return errBadRequest.Withf("%q: Parse error (expected a bool value)", tag)
		}
		v.SetBool(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := strconv.ParseInt(value[0], 10, 64)
		if err != nil {
			return errBadRequest.Withf("%q: Parse error (expected a int value)", tag)
		}
		v.SetInt(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value, err := strconv.ParseUint(value[0], 10, 64)
		if err != nil {
			return errBadRequest.Withf("%q: Parse error (expected a uint value)", tag)
		}
		v.SetUint(value)
	case reflect.Struct:
		switch v.Type() {
		case typeTime:
			t := new(time.Time)
			if len(value) > 0 {
				quoted := strconv.Quote(value[0])
				if err := t.UnmarshalJSON([]byte(quoted)); err != nil {
					return errBadRequest.Withf("%q: Parse error (expected a time value)", tag)
				}
			}
			v.Set(reflect.ValueOf(t).Elem())
		default:
			return errBadRequest.Withf("%q: unsupported type (%q)", tag, v.Type())
		}
	case reflect.Slice:
		// We only support string slices
		if v.Type().Elem().Kind() == reflect.String {
			v.Set(reflect.ValueOf(value))
			return nil
		}
		fallthrough
	default:
		return errBadRequest.Withf("%q: unsupported kind (%q)", tag, v.Kind())
	}

	// Return success
	return nil
}
