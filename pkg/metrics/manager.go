package metrics

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/mutablelogic/go-server/pkg/metrics/schema"
)

////////////////////////////////////////////////////////////////////////
// TYPES

// MetricManager allows a number of metric families to be registered
type MetricManager struct {
	sync.RWMutex
	names    map[string]bool
	families []*schema.Family
}

////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New() *MetricManager {
	manager := new(MetricManager)
	manager.names = make(map[string]bool, 10)
	manager.families = make([]*schema.Family, 0, 10)
	return manager
}

////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m *MetricManager) MarshalJSON() ([]byte, error) {
	m.RLock()
	defer m.RUnlock()

	data := struct {
		Families []*schema.Family `json:"families"`
	}{
		Families: m.families,
	}
	return json.Marshal(data)
}

func (m *MetricManager) String() string {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (m *MetricManager) RegisterFamily(name string, opts ...schema.Opt) (*schema.Family, error) {
	m.Lock()
	defer m.Unlock()

	// Check for duplicate name
	if _, ok := m.names[name]; ok {
		return nil, httpresponse.ErrConflict.Withf("Duplicate metric family name: %q", name)
	}

	// Create a new metric family
	family, err := schema.NewMetricFamily(name, opts...)
	if err != nil {
		return nil, err
	}

	// Register the family
	m.names[name] = true
	m.families = append(m.families, family)

	// Return success
	return family, nil
}

func (m *MetricManager) Write(w io.Writer) error {
	m.RLock()
	defer m.RUnlock()

	for i, family := range m.families {
		if i > 0 {
			if _, err := w.Write([]byte("\n")); err != nil {
				return err
			}
		}
		if err := family.Write(w); err != nil {
			return err
		}
	}
	return nil
}
