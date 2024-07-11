/* implements a dependency graph algorithm */
package dep

import (
	"fmt"
	"slices"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Node any

type node struct {
	Node
	edges []Node
}

type dep struct {
	// Node depends on several other nodes
	edges map[Node][]Node
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewDep() *dep {
	dep := new(dep)
	dep.edges = make(map[Node][]Node)
	return dep
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (d *dep) String() string {
	str := "dep{\n"
	for a, b := range d.edges {
		str += fmt.Sprint("  ", a) + " -> ["
		for i, c := range b {
			if i > 0 {
				str += ", "
			}
			str += fmt.Sprint(c)
		}
		str += "]\n"
	}
	str += "}"
	return str
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Add a node a which depends on several other nodes b
func (d *dep) AddNode(a Node, b ...Node) {
	if _, exists := d.edges[a]; exists {
		d.edges[a] = append(d.edges[a], b...)
	} else {
		d.edges[a] = b
	}
	for _, c := range b {
		if _, exists := d.edges[c]; !exists {
			d.edges[c] = []Node{}
		}
	}
}

// Resolve returns the dependency order for a node
func (d *dep) Resolve(n Node) ([]Node, error) {
	resolved, _, err := d.resolve(n, nil, nil)
	return resolved, err
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Node a depends on several other nodes b
func (d *dep) resolve(n Node, resolved []Node, unresolved []Node) ([]Node, []Node, error) {
	var err error

	unresolved = append(unresolved, n)
	deps := d.edges[n]
	for _, c := range deps {
		if !slices.Contains(resolved, c) {
			if slices.Contains(unresolved, c) {
				return nil, nil, fmt.Errorf("circular dependency detected: %v -> %v", n, c)
			}
			resolved, unresolved, err = d.resolve(c, resolved, unresolved)
			if err != nil {
				return nil, nil, err
			}
		}
	}
	return append(resolved, n), remove(unresolved, n), nil
}

// Remove a node from a list of nodes and return the new list
func remove(nodes []Node, n Node) []Node {
	i := slices.Index(nodes, n)
	if i < 0 {
		return nodes
	}
	return append(nodes[:i], nodes[i+1:]...)
}
