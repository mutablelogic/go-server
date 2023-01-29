package types

import "strings"

/////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	domainSep = "."
)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// DomainFqn returns a fully-qualified value, which includes a trailing dot
func DomainFqn(value ...string) string {
	var result []string
	if len(value) == 0 {
		return ""
	}
	for _, v := range value {
		result = append(result, strings.Trim(v, domainSep))
	}
	return strings.Join(result, domainSep) + domainSep
}

// DomainUnfqn remove final domain separator and domain
func DomainUnfqn(value, domain string) string {
	return strings.Trim(strings.Trim(strings.Trim(value, domainSep), strings.Trim(domain, domainSep)), domainSep)
}

// DomainInZone returns unqualified domain name if it is in the zone
// or an empty string if it is not.
func DomainInZone(value, domain string) string {
	if strings.HasSuffix(value, domainSep+domain) {
		return DomainUnfqn(value, domain)
	} else {
		return ""
	}
}
