package handler

import (
	"context"

	// Packages
	server "github.com/mutablelogic/go-server"
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Register(ctx context.Context, router server.HTTPRouter, prefix string, manager *pgqueue.Manager) {
	registerTicker(ctx, router, prefix, manager)
}
