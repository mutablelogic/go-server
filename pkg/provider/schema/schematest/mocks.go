// Package schematest provides shared mock implementations of [schema.Resource]
// and [schema.ResourceInstance] for use in tests.
package schematest

import (
	"context"

	// Packages
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Resource is a minimal [schema.Resource] stub with a configurable name.
type Resource struct {
	N string // Name returned by Name()
}

// ResourceInstance is a minimal no-op [schema.ResourceInstance] stub.
type ResourceInstance struct {
	N  string // Instance name returned by Name()
	RN string // Resource name returned by Resource().Name()
}

var _ schema.Resource = (*Resource)(nil)
var _ schema.ResourceInstance = (*ResourceInstance)(nil)

///////////////////////////////////////////////////////////////////////////////
// RESOURCE

func (m *Resource) Name() string                                  { return m.N }
func (m *Resource) Schema() []schema.Attribute                    { return nil }
func (m *Resource) New(_ string) (schema.ResourceInstance, error) { return nil, nil }

///////////////////////////////////////////////////////////////////////////////
// RESOURCE INSTANCE

func (m *ResourceInstance) Name() string                    { return m.N }
func (m *ResourceInstance) Resource() schema.Resource       { return &Resource{N: m.RN} }
func (m *ResourceInstance) References() []string            { return nil }
func (m *ResourceInstance) Destroy(_ context.Context) error { return nil }
func (m *ResourceInstance) Read(_ context.Context) (schema.State, error) {
	return nil, nil
}
func (m *ResourceInstance) Validate(_ context.Context, _ schema.State, _ schema.Resolver) (any, error) {
	return nil, nil
}
func (m *ResourceInstance) Plan(_ context.Context, _ any) (schema.Plan, error) {
	return schema.Plan{}, nil
}
func (m *ResourceInstance) Apply(_ context.Context, _ any) error {
	return nil
}
