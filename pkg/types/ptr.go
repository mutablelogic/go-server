package types

import (
	"strings"
	"time"
)

// StringPtr returns a pointer to a string
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

// PtrString returns a string from a pointer
func PtrString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// BoolPtr returns a pointer to a bool
func BoolPtr(v bool) *bool {
	return &v
}

// Int32Ptr returns a pointer to a int32
func Int32Ptr(v int32) *int32 {
	return &v
}

// PtrInt32 returns a int32 from a pointer
func PtrInt32(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

// Uint64Ptr returns a pointer to a uint64
func Uint64Ptr(v uint64) *uint64 {
	return &v
}

// PtrUint64 returns a uint64 from a pointer
func PtrUint64(v *uint64) uint64 {
	if v == nil {
		return 0
	}
	return *v
}

// Int64Ptr returns a pointer to a int64
func Int64Ptr(v int64) *int64 {
	return &v
}

// PtrInt64 returns a int64 from a pointer
func PtrInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

// PtrBool returns a bool from a pointer
func PtrBool(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

// PtrTime returns a pointer to a time.Time
func TimePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// PtrTime returns a time.Time from a pointer
func PtrTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

// DurationPtr returns a pointer to a time.Duration
func DurationPtr(v time.Duration) *time.Duration {
	return &v
}

// PtrDuration returns a duration from a pointer
func PtrDuration(v *time.Duration) time.Duration {
	if v == nil {
		return 0
	}
	return *v
}

// Float64Ptr returns a pointer to a float64
func Float64Ptr(v float64) *float64 {
	return &v
}

// PtrFloat64 returns a float64 from a pointer
func PtrFloat64(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}
