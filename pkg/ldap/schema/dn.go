package schema

import (
	// Packages
	ldap "github.com/go-ldap/ldap/v3"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type DN ldap.DN

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewDN(v string) (*DN, error) {
	if dn, err := ldap.ParseDN(v); err != nil {
		return nil, err
	} else {
		return (*DN)(dn), nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (dn *DN) String() string {
	return (*ldap.DN)(dn).String()
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (dn *DN) AncestorOf(other *DN) bool {
	return (*ldap.DN)(dn).AncestorOf((*ldap.DN)(other))
}

func (dn *DN) Join(other *DN) *DN {
	if other == nil {
		return dn
	}
	return &DN{
		RDNs: append((*ldap.DN)(dn).RDNs, (*ldap.DN)(other).RDNs...),
	}
}
