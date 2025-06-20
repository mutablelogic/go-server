package schema

import (
	"bytes"
	"fmt"
	"io"
	"regexp"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type Sample struct {
	// Name of the sample
	Name string `json:"suffix,omitempty"`

	// Labels are key-value pairs
	Labels []string `json:"labels,omitempty"`

	// Metric value
	Float *float64 `json:"float,omitempty"`
	Int   *int64   `json:"int,omitempty"`

	// Timestamp in seconds since epoch
	Timestamp *float64 `json:"timestamp,omitempty"`
}

/////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	reSampleName = regexp.MustCompile("^[a-zA-Z_:][a-zA-Z0-9_:]*$")
	reLabelName  = regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$")
)

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new float sample with no timestamp
func NewFloat(name string, value float64, labels ...any) (*Sample, error) {
	return newSample(name, types.Float64Ptr(value), nil, labels...)
}

// Create a new integer sample with no timestamp
func NewInt(name string, value int64, labels ...any) (*Sample, error) {
	return newSample(name, nil, types.Int64Ptr(value), labels...)
}

// Gauge type provides a changeable value
func NewGauge(value float64, labels ...any) (*Sample, error) {
	return NewFloat("", value, labels...)
}

// Counter type provides a steadily increasing value
func NewCounter(total float64, labels ...any) (*Sample, error) {
	return NewFloat("total", total, labels...)
}

// State type provides an on/off value
func NewState(value bool, labels ...any) (*Sample, error) {
	boolToInt := func(v bool) int64 {
		if v {
			return 1
		} else {
			return 0
		}
	}
	return NewInt("", boolToInt(value), labels...)
}

// Info type
func NewInfo(labels ...any) (*Sample, error) {
	return NewInt("info", 1, labels...)
}

func newSample(name string, fv *float64, iv *int64, labels ...any) (*Sample, error) {
	if name != "" {
		if !reSampleName.MatchString(name) {
			return nil, httpresponse.ErrBadRequest.Withf("Invalid name: %q", name)
		}
	}

	labelset, err := LabelSet(labels...)
	if err != nil {
		return nil, err
	}

	return &Sample{
		Name:   name,
		Float:  fv,
		Int:    iv,
		Labels: labelset,
	}, nil
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (s *Sample) Write(w io.Writer, prefix string) error {
	var buf bytes.Buffer

	buf.WriteString(prefix)
	if s.Name != "" {
		buf.WriteString("_" + s.Name)
	}

	if len(s.Labels) > 0 {
		buf.WriteString("{")
		for i, label := range s.Labels {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(label)
		}
		buf.WriteString("}")
	}

	buf.WriteString(" ")
	if s.Float != nil {
		buf.WriteString(fmt.Sprint(*s.Float))
	} else if s.Int != nil {
		buf.WriteString(fmt.Sprint(*s.Int))
	} else {
		buf.WriteString("0")
	}
	buf.WriteString("\n")

	// Output buffer
	if _, err := buf.WriteTo(w); err != nil {
		return err
	} else {
		return nil
	}
}
