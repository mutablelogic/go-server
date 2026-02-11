package types

import (
	"strings"
	"time"
)

// Ptr returns a pointer to a value
func Ptr[T any](v T) *T {
	return &v
}

// Value returns a value from a pointer, or a zero value if the pointer is nil
func Value[T any](v *T) T {
	var zero T
	if v == nil {
		return zero
	}
	return *v
}

// StringPtr returns a pointer to a string.
//
// Deprecated: Use [Ptr] instead.
func StringPtr(s string) *string {
	return &s
}

// TrimStringPtr returns a pointer to a white-space trimmed string, or nil
// if the string is empty
func TrimStringPtr(s *string) *string {
	if s == nil {
		return nil
	} else if v := strings.TrimSpace(*s); v == "" {
		return nil
	} else {
		return &v
	}
}

// PtrString returns a string from a pointer.
//
// Deprecated: Use [Value] instead.
func PtrString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// BoolPtr returns a pointer to a bool.
//
// Deprecated: Use [Ptr] instead.
func BoolPtr(v bool) *bool {
	return &v
}

// Int32Ptr returns a pointer to an int32.
//
// Deprecated: Use [Ptr] instead.
func Int32Ptr(v int32) *int32 {
	return &v
}

// PtrInt32 returns an int32 from a pointer.
//
// Deprecated: Use [Value] instead.
func PtrInt32(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

// Uint64Ptr returns a pointer to a uint64.
//
// Deprecated: Use [Ptr] instead.
func Uint64Ptr(v uint64) *uint64 {
	return &v
}

// PtrUint64 returns a uint64 from a pointer.
//
// Deprecated: Use [Value] instead.
func PtrUint64(v *uint64) uint64 {
	if v == nil {
		return 0
	}
	return *v
}

// Int64Ptr returns a pointer to an int64.
//
// Deprecated: Use [Ptr] instead.
func Int64Ptr(v int64) *int64 {
	return &v
}

// PtrInt64 returns an int64 from a pointer.
//
// Deprecated: Use [Value] instead.
func PtrInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

// PtrBool returns a bool from a pointer.
//
// Deprecated: Use [Value] instead.
func PtrBool(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

// TimePtr returns a pointer to a time.Time, or nil if zero.
//
// Deprecated: Use [Ptr] instead (note: Ptr does not nil-on-zero).
func TimePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// PtrTime returns a time.Time from a pointer.
//
// Deprecated: Use [Value] instead.
func PtrTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

// DurationPtr returns a pointer to a time.Duration.
//
// Deprecated: Use [Ptr] instead.
func DurationPtr(v time.Duration) *time.Duration {
	return &v
}

// PtrDuration returns a duration from a pointer.
//
// Deprecated: Use [Value] instead.
func PtrDuration(v *time.Duration) time.Duration {
	if v == nil {
		return 0
	}
	return *v
}

// Float64Ptr returns a pointer to a float64.
//
// Deprecated: Use [Ptr] instead.
func Float64Ptr(v float64) *float64 {
	return &v
}

// PtrFloat64 returns a float64 from a pointer.
//
// Deprecated: Use [Value] instead.
func PtrFloat64(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}
