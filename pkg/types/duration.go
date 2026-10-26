package types

import (
	jsontext "encoding/json/jsontext"
	json "encoding/json/v2"
	"errors"
	"fmt"
	"math"
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
// string, such as "5s", which can start with days and weeks, which are 24 and
// 168 hours, such as "7d", "1w2d" or "1d12h"
func ParseDuration(value string) (time.Duration, error) {
	if n, err := strconv.ParseInt(value, 10, 64); err == nil {
		return time.Duration(n), nil
	}

	// The sign, and then the days and weeks, which are numbers, which can be
	// fractions, followed by d or w
	s, negative := value, false
	if s != "" && (s[0] == '-' || s[0] == '+') {
		s, negative = s[1:], s[0] == '-'
	}
	var days float64
	parsed := false
	for {
		i := 0
		for i < len(s) && (s[i] >= '0' && s[i] <= '9' || s[i] == '.') {
			i++
		}
		if i == 0 || i == len(s) || (s[i] != 'd' && s[i] != 'w') {
			break
		}
		n, err := strconv.ParseFloat(s[:i], 64)
		if err != nil {
			return 0, fmt.Errorf("invalid duration %q", value)
		}
		if s[i] == 'w' {
			n *= 7
		}
		days, s, parsed = days+n, s[i+1:], true
	}
	if !parsed {
		return time.ParseDuration(value)
	}

	// The rest, such as 12h, which has no sign of its own
	var rest time.Duration
	if s != "" {
		if s[0] == '-' || s[0] == '+' {
			return 0, fmt.Errorf("invalid duration %q", value)
		}
		var err error
		if rest, err = time.ParseDuration(s); err != nil {
			return 0, fmt.Errorf("invalid duration %q", value)
		}
	}
	total := days*float64(24*time.Hour) + float64(rest)
	if total > float64(math.MaxInt64) {
		return 0, fmt.Errorf("invalid duration %q: too long", value)
	}
	if negative {
		return -time.Duration(total), nil
	}
	return time.Duration(total), nil
}
