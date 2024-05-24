package provider

import (

	// Packages
	hcl "github.com/mutablelogic/go-hcl/pkg/block"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Set attribute on a block. If the attribute is a resource, then it is
// bound rather than set - then set once the resources have been created
func (self *provider) Set(block hcl.Block, label hcl.Label, value any) error {
	name := block.Name()
	plugin, exists := self.plugin[name]
	if !exists {
		return ErrNotFound.Withf("%q", name)
	} else {
		return plugin.Meta.Set(block, label, value)
	}
}
