package certmanager

import (
	// Packages
	server "github.com/mutablelogic/go-server"
)

////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	X509Name    `hcl:"x509_name" description:"X509 name for certificate"`
	CertStorage CertStorage `hcl:"cert_storage" description:"Certificate storage"`
}

type X509Name struct {
	OrganizationalUnit string `hcl:"organizational_unit,omitempty" description:"X509 Organizational Unit"`
	Organization       string `hcl:"organization" description:"X509 Organization"`
	Locality           string `hcl:"locality,omitempty"  description:"X509 Locality"`
	Province           string `hcl:"province,omitempty"  description:"X509 Province"`
	Country            string `hcl:"country,omitempty"  description:"X509 Country"`
	StreetAddress      string `hcl:"street_address,omitempty"  description:"X509 Street Address"`
	PostalCode         string `hcl:"postal_code,omitempty"  description:"X509 Postal Code"`
}

// Check interfaces are satisfied
var _ server.Plugin = Config{}

////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName = "certmanager"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Name returns the name of the service
func (Config) Name() string {
	return defaultName
}

// Description returns the description of the service
func (Config) Description() string {
	return "certfiicate manager service"
}

// Create a new task from the configuration
func (c Config) New() (server.Task, error) {
	return New(c)
}
