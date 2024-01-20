package provider

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	// Packages
	hcl "github.com/mutablelogic/go-hcl/pkg/block"
	ctx "github.com/mutablelogic/go-server/pkg/context"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Node struct {
	Resource hcl.Resource
	Context  context.Context
	Cancel   context.CancelFunc
	wg       sync.WaitGroup
}

type NodeList []*Node

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Create resources from blocks
func (self *provider) Run(parent context.Context) error {
	var result error

	// Nothing to run
	if len(self.block) == 0 {
		return ErrBadParameter.With("empty configuration")
	}

	// Create the nodes
	var nodes NodeList
	for key, block := range self.block {
		if node, err := self.nodeForBlock(block); err != nil {
			result = errors.Join(result, fmt.Errorf("%w (in %s)", err, key))
			continue
		} else {
			nodes = append(nodes, node)
		}
	}

	// If there are errors, then return them
	if result != nil {
		return result
	}

	// TODO: These should be created in the order to satisfy the dependencies

	// Set loggers
	for _, node := range nodes {
		if logger, ok := node.Resource.(Logger); ok {
			self.logger = append(self.logger, logger)
		}
	}

	// Run the tasks in parallel here, using the parent to trigger
	// cancellation
	for _, node := range nodes {
		node.wg.Add(1)
		go func(node *Node) {
			defer node.wg.Done()
			self.Print(parent, "Running ", ctx.NameLabel(node.Context))
			if err := node.Resource.Run(node.Context); err != nil {
				result = errors.Join(result, err)
			}
		}(node)
	}

	// Wait for cancellation from parent
	<-parent.Done()

	// In reverse order, cancel each
	_ = sort.Reverse(nodes)
	for _, node := range nodes {
		self.Print(parent, "Stopping ", ctx.NameLabel(node.Context))
		node.Cancel()
		node.wg.Wait()
	}

	// Return any errors
	return result
}

///////////////////////////////////////////////////////////////////////////////
// NodeList Methods

func (n NodeList) Len() int {
	return len(n)
}

func (n NodeList) Less(i, j int) bool {
	return false
}

func (n NodeList) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (self *provider) nodeForBlock(block hcl.Block) (*Node, error) {
	parent, cancel := context.WithCancel(context.Background())
	child := ctx.WithNameLabel(ctx.WithProvider(parent, self), block.Name(), self.labelForBlock(block))
	resource, err := block.New(child)
	if err != nil {
		return nil, err
	}
	return &Node{
		Resource: resource,
		Context:  child,
		Cancel:   cancel,
	}, nil
}
