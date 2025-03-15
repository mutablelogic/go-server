package types

import (
	"regexp"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	uuidTwoBytes = `([0-9a-fA-F]{4})`
	uuid         = uuidTwoBytes + uuidTwoBytes + `-` + uuidTwoBytes + `-` + uuidTwoBytes + `-` + uuidTwoBytes + `-` + uuidTwoBytes + uuidTwoBytes + uuidTwoBytes
)

var (
	reUUID = regexp.MustCompile(`^` + uuid + `$`)
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return true if the string is a valid UUID
func IsUUID(id string) bool {
	return reUUID.MatchString(id)
}

// Return two-byte parts of a UUID, or nil if the UUID is invalid.
// The function will always return eight sets if the UUID is valid.
func UUIDSplit(id string) []string {
	parts := reUUID.FindStringSubmatch(strings.ToLower(id))
	if len(parts) > 0 {
		return parts[1:]
	}
	return nil
}
