// Package httphandler has the broadcaster's HTTP API, which streams its
// changes to clients
package httphandler

import (
	_ "embed"

	// Packages
	broadcaster "github.com/mutablelogic/go-server/broadcaster"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	openapi "github.com/mutablelogic/go-server/pkg/openapi"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

//go:embed README.md
var readme []byte

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// RegisterHandlers registers the broadcaster's API with the router, which
// streams its changes at changes, with the "Changes" tag
func RegisterHandlers(b *broadcaster.Broadcaster, router *httprouter.Router) error {
	return changeHandlers(b, router, openapi.ParseMarkdown(readme))
}
