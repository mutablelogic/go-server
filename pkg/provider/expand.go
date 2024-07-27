package provider

import "regexp"

var (
	reEval = regexp.MustCompile(`\$\{([^\}]+)\}|\$([a-zA-Z0-9_\.]+)`)
)

// Expand replaces ${var} or $var in the string based on the mapping function.
// For example, os.ExpandEnv(s) is equivalent to os.Expand(s, os.Getenv).
func Expand(s string, mapping func(string) string) string {
	return reEval.ReplaceAllStringFunc(s, func(s string) string {
		if s[1] == '{' {
			return mapping(s[2 : len(s)-1])
		} else {
			return mapping(s[1:])
		}
	})
}
