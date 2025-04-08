package schema

import (
	"bytes"
	"fmt"
	"io"
	"unicode"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type Sample struct {
	// Name of the sample
	Name string `json:"name"`

	// Labels are key-value pairs
	Labels []string `json:"labels,omitempty"`

	// Metric value
	Float *float64 `json:"value,omitempty"`
	Int   *int64   `json:"value,omitempty"`

	// Timestamp in seconds since epoch
	Timestamp *float64 `json:"timestamp,omitempty"`
}

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
		if !stringContains(name, func(i int, r rune) bool {
			if i == 0 {
			} else {
				return unicode.IsLower(r)
			}
		}) {
			return nil, httpresponse.ErrBadRequest.Withf("Invalid name: %q", name)
		}
	}

	labelset, err := labelSet(labels...)
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

func (s *Sample) Write(prefix string, w io.Writer) error {
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

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func labelSet(labels ...any) ([]string, error) {
	labelset := make([]string, 0, len(labels)>>1)
	for i := 0; i < len(labels); i += 2 {
		key, ok := labels[i].(string)
		if !ok || key == "" {
			return nil, httpresponse.ErrBadRequest.Withf("Invalid key: %v", labels[i])
		} else if !types.StringContains(key, func(r rune) bool {
			return unicode.IsLower(r)
		}) {
			return nil, httpresponse.ErrBadRequest.Withf("Invalid key: %q", key)
		}
		labelset = append(labelset, fmt.Sprintf("%v=%q", key, labels[i+1]))
	}
	return labelset, nil
}
