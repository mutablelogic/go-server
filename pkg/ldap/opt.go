package ldap

import (
	"net/url"

	// Packages
	"github.com/mutablelogic/go-server/pkg/httpresponse"
	schema "github.com/mutablelogic/go-server/pkg/ldap/schema"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type opt struct {
	url        *url.URL
	user       string
	pass       string
	dn         *schema.DN
	skipverify bool
	users      *schema.ObjectType
	groups     *schema.ObjectType
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

func WithUserSchema(dn, field string, classes ...string) Opt {
	return func(o *opt) error {
		if ot, err := schema.NewObjectType(dn, field, classes...); err != nil {
			return err
		} else {
			o.users = ot
		}
		return nil
	}
}

func WithGroupSchema(dn, field string, classes ...string) Opt {
	return func(o *opt) error {
		if ot, err := schema.NewObjectType(dn, field, classes...); err != nil {
			return err
		} else {
			o.groups = ot
		}
		return nil
	}
}

func WithUrl(v string) Opt {
	return func(o *opt) error {
		if u, err := url.Parse(v); err != nil {
			return err
		} else {
			o.url = u
		}
		return nil
	}
}

func WithBaseDN(v string) Opt {
	return func(o *opt) error {
		if v == "" {
			return nil
		} else if bdn, err := schema.NewDN(v); err != nil {
			return httpresponse.ErrBadRequest.With("DN is invalid: ", err)
		} else {
			o.dn = bdn
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

func WithSkipVerify() Opt {
	return func(o *opt) error {
		o.skipverify = true
		return nil
	}
}
