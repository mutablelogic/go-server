package httprequest

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	tagStruct = "json"
)

type Unmarshaller interface {
	// Convert a string into a custom type
	UnmarshalJSON([]byte) error
}

var (
	typeUnmarshal          = reflect.TypeOf((*Unmarshaller)(nil)).Elem()
	typeDuration           = reflect.TypeOf(time.Duration(0))
	typeMultiPartFilePtr   = reflect.TypeOf((*multipart.FileHeader)(nil))
	typeSliceMultiPartFile = reflect.TypeOf([]*multipart.FileHeader{})
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Query copies the query parameters from a request into a structure, which
// should be a pointer to a struct. The struct should have json tags. No attempt
// is made to validate the query parameters beyond the conversion from string
// or []string into the appropriate type. It can use the MarshalJSON interface
// to convert a string into a custom type.
func Query(v any, req url.Values) error {
	return mapFields(v, func(key string, value reflect.Value) error {
		json, exists := req[key]
		if !exists || len(json) == 0 {
			// Won't mess with file fields - they are done in a separate step
			if value.Type() != typeMultiPartFilePtr && value.Type() != typeSliceMultiPartFile {
				setZeroValue(value)
			}
			return nil
		}
		return setValue(value, json)
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Map fields iteratres over a structure v and calls fn for each field
// with the json tag name and the value of the field
func mapFields(v any, fn func(key string, value reflect.Value) error) error {
	var result error

	// Convert v to a struct
	rv, err := toStruct(v)
	if err != nil {
		return err
	}

	// Enumerate fields - call fn for each field
	fields := reflect.VisibleFields(rv.Type())
	if len(fields) == 0 {
		return ErrBadParameter.With("v has no public fields")
	}
	for _, field := range fields {
		// Get tag and field value - ignore if the field cannot be set
		tag, value := jsonName(field), rv.FieldByName(field.Name)
		if tag == "" || !value.CanSet() {
			continue
		} else if err := fn(tag, value); err != nil {
			result = errors.Join(result, fmt.Errorf("%s: %w (from %q)", field.Name, err, tag))
		}
	}

	// Return any errors
	return result
}

// Set a struct field to a zero value
func setZeroValue(v reflect.Value) {
	v.Set(reflect.Zero(v.Type()))
}

// Set a struct field to a value
func setValue(v reflect.Value, json []string) error {
	// If the value is a pointer, then create a new value and set it, and deference
	if v.Kind() == reflect.Ptr {
		v.Set(reflect.New(v.Type().Elem()))
		v = v.Elem()
	}

	// Set the value based on the kind of the field
	switch v.Kind() {
	case reflect.Slice:
		switch v.Type().Elem().Kind() {
		case reflect.String:
			v.Set(reflect.ValueOf(json))
		default:
			if arr, err := toArray(json, v.Type().Elem()); err != nil {
				return err
			} else {
				v.Set(arr)
			}
		}
	default:
		value, err := toScalar(json[0], v.Type())
		if err != nil {
			return err
		}
		v.Set(value)
	}

	// Return success
	return nil
}

// Set a struct field to a value
// TODO: We should support a slice of files
func setFile(v reflect.Value, files []*multipart.FileHeader) error {
	switch v.Type() {
	case typeMultiPartFilePtr:
		v.Set(reflect.ValueOf(files[0]))
		return nil
	default:
		return fmt.Errorf("unsupported type %q", v.Type())
	}
}

// Create an array of values
func toArray(json []string, t reflect.Type) (reflect.Value, error) {
	arr := reflect.MakeSlice(reflect.SliceOf(t), len(json), len(json))
	for i := range json {
		value, err := toScalar(json[i], t)
		if err != nil {
			return reflect.Value{}, err
		}
		arr.Index(i).Set(value)
	}
	return arr, nil
}

// Convert a string value to a scalar value
func toScalar(json string, t reflect.Type) (reflect.Value, error) {
	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf(json).Convert(t), nil
	case reflect.Bool:
		if v, err := strconv.ParseBool(json); err != nil {
			return reflect.Value{}, err
		} else {
			return reflect.ValueOf(v).Convert(t), nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Special case for time.Duration
		if t == typeDuration {
			if v, err := time.ParseDuration(json); err != nil {
				return reflect.Value{}, err
			} else {
				return reflect.ValueOf(v).Convert(t), nil
			}
		} else if v, err := strconv.ParseInt(json, 10, 64); err != nil {
			return reflect.Value{}, err
		} else {
			return reflect.ValueOf(v).Convert(t), nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v, err := strconv.ParseUint(json, 10, 64); err != nil {
			return reflect.Value{}, err
		} else {
			return reflect.ValueOf(v).Convert(t), nil
		}
	case reflect.Float32, reflect.Float64:
		if v, err := strconv.ParseFloat(json, 64); err != nil {
			return reflect.Value{}, err
		} else {
			return reflect.ValueOf(v).Convert(t), nil
		}
	case reflect.Struct:
		// null returns a zero value
		if json == "null" {
			return reflect.Zero(t), nil
		}
		// For a struct to be parsable, if needs to implement Unmarshaller interface
		if reflect.PointerTo(t).Implements(typeUnmarshal) {
			v := reflect.New(t)
			json = strconv.Quote(json)
			if err := v.Interface().(Unmarshaller).UnmarshalJSON([]byte(json)); err != nil {
				return reflect.Value{}, err
			}
			return v.Elem(), nil
		}
		return reflect.Value{}, fmt.Errorf("unsupported type %q", t)
	default:
		return reflect.Value{}, fmt.Errorf("unsupported kind %q", t.Kind())
	}
}

// Convert an any value to a struct, return error if v is the wrong type
func toStruct(v any) (reflect.Value, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return reflect.Value{}, ErrBadParameter.With("v must be a pointer")
	} else {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return reflect.Value{}, ErrBadParameter.With("v must be a pointer to a struct")
	}
	return rv, nil
}

// Return the json field name for a struct field, or empty string if the field
// should be ignored
func jsonName(field reflect.StructField) string {
	tag := field.Tag.Get(tagStruct)
	if tag == "-" {
		return ""
	}
	if fields := strings.Split(tag, ","); len(fields) > 0 && fields[0] != "" {
		return fields[0]
	}
	return field.Name
}
