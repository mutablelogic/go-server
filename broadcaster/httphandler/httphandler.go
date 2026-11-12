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

// Tag is the tag of the broadcaster's API, which a server with tag groups
// adds to one of them, as Redoc doesn't show a tag which isn't in a group,
// once there are groups, such as with:
//
//	router.Spec().AddTagGroup("Server", httphandler.Tag)
const Tag = "Changes"

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// RegisterHandlers registers the broadcaster's API with the router, which
// streams its changes at changes, with Tag
func RegisterHandlers(b *broadcaster.Broadcaster, router *httprouter.Router) error {
	return changeHandlers(b, router, openapi.ParseMarkdown(readme))
}
