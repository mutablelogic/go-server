package tokenauth

import (
	"context"
	"os"
	"path/filepath"
	"time"

	// Namespace imports
	iface "github.com/mutablelogic/go-server"
	ctx "github.com/mutablelogic/go-server/pkg/context"
	task "github.com/mutablelogic/go-server/pkg/task"
)

/////////////////////////////////////////////////////////////////////
// TYPES

// Plugin for tokenauth for storing tokens in a file
type Plugin struct {
	task.Plugin
	Path_  string        `json:"path,omitempty"`
	File_  string        `json:"file,omitempty"`
	Delta_ time.Duration `json:"delta,omitempty"`
}

/////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// Scope for reading tokens
	ScopeRead = "github.com/mutablelogic/go-server/tokenauth:read"

	// Scope for updating, deleting and creating tokens
	ScopeWrite = "github.com/mutablelogic/go-server/tokenauth:write"
)

const (
	AdminToken     = "admin"
	defaultName    = "tokenauth"
	defaultFileExt = ".json"
	defaultLength  = 32
	defaultDelta   = time.Second * 30
)

var (
	adminScopes = []string{ScopeRead, ScopeWrite}
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func (p Plugin) New(parent context.Context, _ iface.Provider) (iface.Task, error) {
	// Check parameters
	if err := p.HasNameLabel(); err != nil {
		return nil, err
	}

	return NewWithPlugin(p, ctx.NameLabel(parent))
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func WithLabel(label string) Plugin {
	return Plugin{
		Plugin: task.WithLabel(defaultName, label),
	}
}

func (p Plugin) WithPath(path string) Plugin {
	p.Path_ = path
	return p
}

func (p Plugin) Name() string {
	if name := p.Plugin.Name(); name != "" {
		return name
	} else {
		return defaultName
	}
}

func (p Plugin) Path() (string, error) {
	var path = p.Path_

	// Get config directory
	if path == "" {
		if v, err := os.UserConfigDir(); err != nil {
			return "", err
		} else {
			path = filepath.Join(v, p.Name())
		}
	}

	// Make path
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", err
	}

	// Return path
	return path, nil
}

func (p Plugin) File() string {
	if p.File_ == "" {
		return p.Label() + defaultFileExt
	} else {
		return p.File_
	}
}

func (p Plugin) Delta() time.Duration {
	if p.Delta_ <= 0 {
		return defaultDelta
	} else {
		return p.Delta_
	}
}
