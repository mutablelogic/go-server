package dep_test

import (
	"testing"

	dep "github.com/mutablelogic/go-server/pkg/provider/dep"
	"github.com/stretchr/testify/assert"
)

type node struct {
	name string
}

func (n node) String() string {
	return n.name
}

func NewNode(name string) *node {
	return &node{name}
}

func Test_dep_00(t *testing.T) {
	d := dep.NewDep()
	if d == nil {
		t.Fatalf("Expected a new dep, got nil")
	}

	root := NewNode("root")
	a := NewNode("a")
	b := NewNode("b")
	c := NewNode("c")

	// Add nodes a, b and c with dependencies
	d.AddNode(a, b, c)
	d.AddNode(root, a, b, c)

	resolved, err := d.Resolve(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resolved)
}

func Test_dep_01(t *testing.T) {
	assert := assert.New(t)
	d := dep.NewDep()
	if !assert.NotNil(d) {
		t.SkipNow()
	}

	root := NewNode("root")
	a := NewNode("a")
	b := NewNode("b")
	c := NewNode("c")

	// Add nodes a, b and c with dependencies
	d.AddNode(a, b, c)
	d.AddNode(root, a, b, c)
	d.AddNode(b, root)

	_, err := d.Resolve(root)
	if !assert.NoError(err) {
		t.SkipNow()
	}
}
