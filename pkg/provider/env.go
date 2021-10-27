package provider

import (
	"context"
	"fmt"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type envvars struct {
	v map[string]interface{}
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewEnvVars(v map[string]interface{}) Env {
	e := new(envvars)
	e.v = v
	return e
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (e *envvars) GetString(ctx context.Context, key string) (string, error) {
	if v, exists := e.v[key]; exists {
		if v == nil {
			return "", nil
		}
		switch v := v.(type) {
		case string:
			return v, nil
		default:
			return fmt.Sprint(v), nil
		}
	}
	return "", ErrNotFound.With(key)
}

func (p *provider) GetString(ctx context.Context, key string) (string, error) {
	for _, env := range p.env {
		if value, err := env.GetString(ctx, key); err == nil {
			return value, nil
		}
	}
	return "", ErrNotFound.With(key)
}
