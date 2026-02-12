package types

import (
	"strings"
)

const (
	pathSeparator = "/"
)

// Return path with a single '/' separator at the beginning and no
// trailing '/' separator
func NormalisePath(path string) string {
	if path == "" {
		return pathSeparator
	}
	path = pathSeparator + strings.TrimLeft(path, pathSeparator)
	if path == pathSeparator {
		return path
	}
	return strings.TrimRight(path, pathSeparator)
}

// Return a normalised joined path
func JoinPath(prefix, path string) string {
	return NormalisePath(strings.TrimRight(prefix, pathSeparator) + pathSeparator + strings.TrimLeft(path, pathSeparator))
}
