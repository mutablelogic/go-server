package client

import (
	"fmt"
	"net/url"

	"github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type opt struct {
	url.Values
}

// An Option to set on the client
type Opt func(*opt) error

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func applyOpts(opts ...Opt) (*opt, error) {
	o := new(opt)
	o.Values = make(url.Values)
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}

////////////////////////////////////////////////////////////////////////////////
// OPTIONS

// Set offset and limit
func WithOffsetLimit(offset uint64, limit *uint64) Opt {
	return func(o *opt) error {
		if offset > 0 {
			o.Set("offset", fmt.Sprint(offset))
		}
		if limit != nil {
			o.Set("limit", fmt.Sprint(*limit))
		}
		return nil
	}
}

func WithForce(v bool) Opt {
	if v {
		return OptSet("force", fmt.Sprint(v))
	} else {
		return OptSet("force", "")
	}
}

func WithStatus(v *string) Opt {
	return OptSet("status", types.PtrString(v))
}

func WithScope(v *string) Opt {
	return OptSet("scope", types.PtrString(v))
}

func OptSet(k, v string) Opt {
	return func(o *opt) error {
		if v == "" {
			o.Del(k)
		} else {
			o.Set(k, v)
		}
		return nil
	}
}
