package ldap

// Packages
import "net/url"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type opt struct {
	url        *url.URL
	user       string
	pass       string
	dn         string
	skipverify bool
}

// Opt represents a function that modifies the options
type Opt func(*opt) error

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func applyOpts(opts ...Opt) (*opt, error) {
	var o opt

	// Apply the options
	for _, fn := range opts {
		if err := fn(&o); err != nil {
			return nil, err
		}
	}

	// Return success
	return &o, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func WithUrl(v string) Opt {
	return func(o *opt) error {
		if v = v; v != "" {
			if u, err := url.Parse(v); err != nil {
				return err
			} else {
				o.url = u
			}
		}
		return nil
	}
}

func WithBaseDN(v string) Opt {
	return func(o *opt) error {
		if v != "" {
			o.dn = v
		}
		return nil
	}
}

func WithUser(v string) Opt {
	return func(o *opt) error {
		if v != "" {
			o.user = v
		}
		return nil
	}
}

func WithPassword(v string) Opt {
	return func(o *opt) error {
		if v != "" {
			o.pass = v
		}
		return nil
	}
}

func WithSkipVerify(v bool) Opt {
	return func(o *opt) error {
		o.skipverify = v
		return nil
	}
}
