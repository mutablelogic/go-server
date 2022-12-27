package unmarshal

import (
	"math"
	"reflect"
	"strconv"
	"time"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	typeTime     = reflect.TypeOf(time.Time{})
	typeDuration = reflect.TypeOf(time.Duration(0))
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Copy across a value from source to destination, converting types as needed
func setValue(dest, src reflect.Value) error {
	// Dereference destination pointer
	if dest.Kind() == reflect.Ptr {
		dest = dest.Elem()
	}
	// Set value
	switch dest.Kind() {
	case reflect.String:
		dest.SetString(src.String())
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if dest.Type() == typeDuration {
			v, err := toDuration(src)
			if err != nil {
				return err
			}
			dest.SetInt(v)
		} else {
			v, err := toInt64(src)
			if err != nil {
				return err
			}
			dest.SetInt(v)
		}
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := toUint64(src)
		if err != nil {
			return err
		}
		dest.SetUint(v)
		return nil
	case reflect.Float32, reflect.Float64:
		v, err := toFloat64(src)
		if err != nil {
			return err
		}
		dest.SetFloat(v)
		return nil
	case reflect.Bool:
		v, err := toBool(src)
		if err != nil {
			return err
		}
		dest.SetBool(v)
		return nil
	case reflect.Struct:
		if dest.Type() == typeTime {
			v, err := toTime(src)
			if err != nil {
				return err
			}
			dest.Set(reflect.ValueOf(v))
			return nil
		}
	}

	// Return error - unsupported type
	return ErrBadParameter.Withf("unsupported type %q", dest.Kind())
}

// Convert arbitary value to time
func toTime(v reflect.Value) (time.Time, error) {
	switch v.Kind() {
	case reflect.String:
		if v, err := time.Parse(time.RFC3339, v.String()); err != nil {
			return time.Time{}, err
		} else {
			return v, nil
		}
	}
	return time.Time{}, ErrBadParameter.Withf("unsupported type %q", v.Kind())
}

// Convert arbitary value to duration
func toDuration(v reflect.Value) (int64, error) {
	switch v.Kind() {
	case reflect.String:
		if v, err := time.ParseDuration(v.String()); err != nil {
			return 0, err
		} else {
			return int64(v), nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(v.Uint()), nil
	}
	return 0, ErrBadParameter.Withf("unsupported type %q", v.Kind())
}

// Convert arbitary value to int64
func toInt64(v reflect.Value) (int64, error) {
	switch v.Kind() {
	case reflect.String:
		if v, err := strconv.ParseInt(v.String(), 10, 64); err != nil {
			return 0, err
		} else {
			return v, nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.Uint() > math.MaxInt64 {
			return 0, ErrBadParameter.Withf("value too large to be converted to signed integer")
		} else {
			return int64(v.Uint()), nil
		}
	case reflect.Float32, reflect.Float64:
		return int64(v.Float()), nil
	default:
		return 0, ErrBadParameter.Withf("unsupported type %q", v.Kind())
	}
}

// Convert arbitary value to uint64
func toUint64(v reflect.Value) (uint64, error) {
	switch v.Kind() {
	case reflect.String:
		if v, err := strconv.ParseUint(v.String(), 10, 64); err != nil {
			return 0, err
		} else {
			return v, nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() < 0 {
			return 0, ErrBadParameter.Withf("negative value cannot be converted to unsigned integer")
		}
		return uint64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint(), nil
	case reflect.Float32, reflect.Float64:
		if v.Float() < 0 {
			return 0, ErrBadParameter.Withf("negative value cannot be converted to unsigned integer")
		}
		return uint64(v.Float()), nil
	default:
		return 0, ErrBadParameter.Withf("unsupported type %q", v.Kind())
	}
}

// Convert arbitary value to float64
func toFloat64(v reflect.Value) (float64, error) {
	switch v.Kind() {
	case reflect.String:
		if v, err := strconv.ParseFloat(v.String(), 64); err != nil {
			return 0, err
		} else {
			return v, nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	default:
		return 0, ErrBadParameter.Withf("unsupported type %q", v.Kind())
	}
}

// Convert arbitary value to bool
func toBool(v reflect.Value) (bool, error) {
	switch v.Kind() {
	case reflect.String:
		if v, err := strconv.ParseBool(v.String()); err != nil {
			return false, err
		} else {
			return v, nil
		}
	case reflect.Bool:
		return v.Bool(), nil
	default:
		return false, ErrBadParameter.Withf("unsupported type %q", v.Kind())
	}
}
