package types

import "time"

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
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
