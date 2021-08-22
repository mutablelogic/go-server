package server

import (
	"context"
)

// Provider provides services to a module
type Provider interface {
	// GetPlugin returns a named plugin or nil if not available
	GetPlugin(context.Context, string) Plugin

	// GetConfig populates yaml config
	GetConfig(context.Context, interface{}) error
}

// Plugin provides handlers to server
type Plugin interface {
	// Run plugin background tasks until cancelled
	Run(context.Context) error
}

// Logger providers a logging interface
type Logger interface {
	Print(context.Context, ...interface{})
	Printf(context.Context, string, ...interface{})
}
