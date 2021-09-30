package provider

import (
	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// ENV IMPLEMENTATION

func (p *provider) GetString(key string) (string, error) {
	for _, env := range p.env {
		if value, err := env.GetString(key); err == nil {
			return value, nil
		}
	}
	return "", ErrNotFound.With(key)
}
