package main

import (
	// Packages
	server "github.com/mutablelogic/go-server"
	ldap "github.com/mutablelogic/go-server/pkg/ldap/config"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func Plugin() server.Plugin {
	return ldap.Config{}
}
