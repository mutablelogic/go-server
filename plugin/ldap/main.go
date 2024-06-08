package main

import (
	// Packages
	server "github.com/mutablelogic/go-server"
	ldap "github.com/mutablelogic/go-server/pkg/handler/ldap"
)

func Plugin() server.Plugin {
	return ldap.Config{}
}
