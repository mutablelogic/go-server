package simplerouter

import (
	"context"
	"fmt"
	"net/http"

	// Packages
	hcl "github.com/mutablelogic/go-hcl/pkg/block"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// logging configuration
type Config struct {
	Services map[string]hcl.Resource
}

// http router instance
type httprouter struct {
	http.ServeMux
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName = "simplerouter"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (c Config) Name() string {
	return defaultName
}

func (c Config) Description() string {
	return "provides simple HTTP routing"
}

func (c Config) New(context.Context) (hcl.Resource, error) {
	self := new(httprouter)
	self.ServeMux = *http.NewServeMux()

	// Return success
	return self, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c Config) String() string {
	str := "<block "
	str += fmt.Sprintf("name=%q", c.Name())
	return str + ">"
}

func (self *httprouter) String() string {
	str := "<" + defaultName
	return str + ">"
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (self *httprouter) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
