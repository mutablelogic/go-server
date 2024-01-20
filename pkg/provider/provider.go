package provider

import (
	"context"
	"errors"
	"fmt"

	// Packages
	hcl "github.com/mutablelogic/go-hcl/pkg/block"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Provider represents all the resources for a given configuration
type provider struct {
	plugin map[string]*Plugin
	block  map[string]hcl.Block

	// All loggers
	logger []Logger
}

// Ensure that provider implements the Provider interface
var _ Provider = (*provider)(nil)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultCapacity = 20
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new tasks object and load plugins
func NewPluginsForPattern(ctx context.Context, patterns ...string) (*provider, error) {
	var result error

	// Create tasks
	self := &provider{
		plugin: make(map[string]*Plugin, defaultCapacity),
		block:  make(map[string]hcl.Block, defaultCapacity),
	}

	for _, pattern := range patterns {
		plugins, err := LoadPluginsForPattern(pattern)
		if err != nil {
			result = errors.Join(result, err)
			continue
		}
		for _, plugin := range plugins {
			if _, exists := self.plugin[plugin.Name]; exists {
				result = errors.Join(result, ErrDuplicateEntry.Withf("%q", plugin.Name))
			} else {
				self.plugin[plugin.Name] = plugin
			}
		}
	}

	// Return success
	return self, result
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (self *provider) String() string {
	str := "<provider"
	for _, plugin := range self.plugin {
		str += "\n  " + fmt.Sprint(plugin)
	}
	for _, block := range self.block {
		str += "\n  " + fmt.Sprint(block)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return a list of paths and plugins
func (self *provider) Plugins() map[string]*hcl.BlockMeta {
	var result = make(map[string]*hcl.BlockMeta, len(self.plugin))
	for _, plugin := range self.plugin {
		result[plugin.Path] = plugin.Meta
	}
	return result
}

// Create a new block from a label, and return the block
func (self *provider) NewBlock(name string, label hcl.Label) (hcl.Block, error) {
	// Compute key
	key := keyForBlock(name, label)
	if _, exists := self.block[key]; exists {
		return nil, ErrDuplicateEntry.Withf("%q", key)
	}

	// Check label
	if !(label.IsZero() || label.IsValid()) {
		return nil, ErrBadParameter.Withf("invalid label %q", label)
	}

	// Get plugin
	plugin, exists := self.plugin[name]
	if !exists {
		return nil, ErrNotFound.Withf("Unable to create block %q", key)
	}

	// Create block
	block, err := plugin.Meta.New(label)
	if err != nil {
		return nil, err
	} else if block == nil {
		return nil, ErrBadParameter.Withf("cannot create block %q", key)
	}

	// Set block
	if _, exists := self.block[key]; exists {
		return nil, ErrDuplicateEntry.Withf("block with key %q", key)
	} else {
		self.block[key] = block
	}

	// Return success
	return block, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func keyForBlock(name string, label hcl.Label) string {
	if label.IsZero() {
		return name
	} else {
		return name + string(hcl.RuneLabelSeparator) + label.String()
	}
}

func (self *provider) labelForBlock(block hcl.Block) hcl.Label {
	plugin, exists := self.plugin[block.Name()]
	if !exists {
		return nil
	}
	return plugin.Meta.GetLabel(block)
}
