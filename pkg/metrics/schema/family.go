package schema

import (
	"bytes"
	"io"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type Family struct {
	// Name of the family
	Name string `json:"name,omitempty"`

	// Help text
	Help string `json:"help,omitempty"`

	// Valid values are "gauge", "counter", "stateset", "info", "histogram", "gaugehistogram", and "summary".
	// If not set, "unknown" is assumed.
	Type string `json:"type"`

	// Unit of measurement, ie "seconds". This is optional.
	Unit string `json:"unit,omitempty"`

	// Samples
	Samples []*Sample `json:"samples,omitempty"`
}

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewMetricFamily returns a new, empty metric family
func NewMetricFamily(name string, opts ...Opt) (*Family, error) {
	opt, err := apply(opts...)
	if err != nil {
		return nil, err
	}

	// Return success
	return &Family{
		Name:    name,
		Type:    opt.typ,
		Help:    opt.help,
		Unit:    opt.unit,
		Samples: opt.samples,
	}, nil
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (f *Family) Write(w io.Writer) error {
	var buf bytes.Buffer
	buf.WriteString("# TYPE " + f.nameWithUnit() + " " + f.Type + "\n")
	if f.Unit != "" {
		buf.WriteString("# UNIT " + f.nameWithUnit() + " " + f.Unit + "\n")
	}
	if f.Help != "" {
		buf.WriteString("# HELP " + f.nameWithUnit() + " " + escapeString(f.Help) + "\n")
	}

	// Write samples
	for _, sample := range f.Samples {
		if err := sample.Write(f.Name, &buf); err != nil {
			return err
		}
	}

	// Output buffer
	if _, err := buf.WriteTo(w); err != nil {
		return err
	} else {
		return nil
	}
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (f *Family) nameWithUnit() string {
	if f.Unit != "" {
		return f.Name + "_" + f.Unit
	} else {
		return f.Name
	}
}
