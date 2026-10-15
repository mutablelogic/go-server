package types

import (
	jsontext "encoding/json/jsontext"
	json "encoding/json/v2"
	"errors"
	"fmt"
	"strconv"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// DurationUnmarshalers decodes a time.Duration from a duration string, such as
// "5s", or integer nanoseconds, as a number or a string. Use it with
// json.WithUnmarshalers when unmarshaling with encoding/json/v2. Other values,
// such as null, are decoded as usual.
var DurationUnmarshalers = json.UnmarshalFromFunc(func(dec *jsontext.Decoder, d *time.Duration) error {
	var value string
	switch dec.PeekKind() {
	case '"':
		token, err := dec.ReadToken()
		if err != nil {
			return err
		}
		value = token.String()
	case '0':
		number, err := dec.ReadValue()
		if err != nil {
			return err
		}
		value = string(number)
	default:
		return errors.ErrUnsupported
	}

	// The error has the JSON Pointer of the value, such as "/timeout", as an
	// error from an unmarshaler isn't given one with v1 options
	duration, err := ParseDuration(value)
	if err != nil {
		return fmt.Errorf("%q: %w", dec.StackPointer(), err)
	}
	*d = duration
	return nil
})

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ParseDuration parses a duration, which is integer nanoseconds, or a duration
// string, such as "5s"
func ParseDuration(value string) (time.Duration, error) {
	if n, err := strconv.ParseInt(value, 10, 64); err == nil {
		return time.Duration(n), nil
	}
	return time.ParseDuration(value)
}
