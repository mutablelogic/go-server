package schema

import (
	"unicode"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type Opt func(*opt) error

type opt struct {
	typ     string
	help    string
	unit    string
	samples []*Sample
}

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func apply(opts ...Opt) (*opt, error) {
	opt := new(opt)
	opt.typ = "unknown"
	for _, fn := range opts {
		if err := fn(opt); err != nil {
			return nil, err
		}
	}
	return opt, nil
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func WithHelp(help string) Opt {
	return func(opt *opt) error {
		opt.help = help
		return nil
	}
}

func WithUnit(unit string) Opt {
	return func(opt *opt) error {
		// Check to make sure all runes are lowercase
		if !stringContains(unit, func(_ int, r rune) bool {
			return unicode.IsLower(r)
		}) {
			return httpresponse.ErrBadRequest.Withf("Invalid unit: %q", unit)
		} else {
			opt.unit = unit
		}
		return nil
	}
}

func WithType(v string) Opt {
	return func(opt *opt) error {
		switch v {
		case "gauge", "counter", "stateset", "info", "summary", "histogram", "gaugehistogram":
			opt.typ = v
		default:
			return httpresponse.ErrBadRequest.Withf("Invalid type: %q", v)
		}
		return nil
	}
}

// With a float64 value
func WithFloat(name string, value float64, labels ...any) Opt {
	return func(opt *opt) error {
		if sample, err := NewFloat(name, value, labels...); err != nil {
			return err
		} else {
			opt.samples = append(opt.samples, sample)
		}
		return nil
	}
}

// With an integer value
func WithInt(name string, value int64, labels ...any) Opt {
	return func(opt *opt) error {
		if sample, err := NewInt(name, value, labels...); err != nil {
			return err
		} else {
			opt.samples = append(opt.samples, sample)
		}
		return nil
	}
}

// Gauge type provides a changeable value
func WithGauge(value float64, labels ...any) Opt {
	return func(opt *opt) error {
		if sample, err := NewGauge(value, labels...); err != nil {
			return err
		} else {
			opt.samples = append(opt.samples, sample)
		}
		return nil
	}
}

// Counter type provides a steadily increasing value
func WithCounter(total float64, labels ...any) Opt {
	return func(opt *opt) error {
		if sample, err := NewCounter(total, labels...); err != nil {
			return err
		} else {
			opt.samples = append(opt.samples, sample)
		}
		return nil
	}
}

// State type provides an on/off value
func WithState(value bool, labels ...any) Opt {
	return func(opt *opt) error {
		if sample, err := NewState(value, labels...); err != nil {
			return err
		} else {
			opt.samples = append(opt.samples, sample)
		}
		return nil
	}
}

// Info type
func WithInfo(labels ...any) Opt {
	return func(opt *opt) error {
		if sample, err := NewInfo(labels...); err != nil {
			return err
		} else {
			opt.samples = append(opt.samples, sample)
		}
		return nil
	}
}
