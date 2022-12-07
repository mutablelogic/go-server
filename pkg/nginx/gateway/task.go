package gateway

import (
	"fmt"
	"net/http"
	"regexp"

	// Package imports
	ctx "github.com/mutablelogic/go-server/pkg/context"
	task "github.com/mutablelogic/go-server/pkg/task"
	plugin "github.com/mutablelogic/go-server/plugin"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type t struct {
	task.Task
	prefix string
	nginx  plugin.Nginx
	scoped bool
}

var (
	reRoot   = regexp.MustCompile(`^/$`)
	reReload = regexp.MustCompile(`^/reload$`)
	reReopen = regexp.MustCompile(`^/reopen$`)
	reTest   = regexp.MustCompile(`^/test$`)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func NewWithPlugin(p Plugin) (*t, error) {
	t := new(t)

	if nginx, ok := p.Task().(plugin.Nginx); !ok || nginx == nil {
		return nil, ErrBadParameter.With("nginx")
	} else {
		t.nginx = nginx
		t.prefix = p.Prefix()
		t.scoped = true
	}

	// Return success
	return t, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t *t) String() string {
	str := "<nginx-gateway"
	if t.prefix != "" {
		str += fmt.Sprintf(" prefix=%q", t.prefix)
	}
	if t.scoped {
		str += fmt.Sprintf(" read_scope=%q write_scope=%q", readScope, writeScope)
	}
	if t.nginx != nil {
		str += fmt.Sprint(" nginx=", t.nginx)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (t *t) Prefix() string {
	return t.prefix
}

func (t *t) RegisterHandlers(router plugin.Router) error {
	// Path: /nginx/v1
	// Methods: GET
	// Scopes: read, write
	// Description: Get nginx status (version, uptime, available and enabled configurations)
	if err := router.AddHandler(t, reRoot, ctx.RequireScope(t.StatusHandler, t.scoped, readScope, writeScope)); err != nil {
		return err
	}

	// Path: /nginx/v1/reopen
	// Methods: POST
	// Scopes: write
	// Description: Reopen nginx files
	if err := router.AddHandler(t, reReopen, ctx.RequireScope(t.ReopenHandler, t.scoped, writeScope), http.MethodPost); err != nil {
		return err
	}

	// Path: /nginx/v1/reload
	// Methods: POST
	// Scopes: write
	// Description: Restart nginx configuration after testing configuration
	if err := router.AddHandler(t, reReload, ctx.RequireScope(t.ReloadHandler, t.scoped, writeScope), http.MethodPost); err != nil {
		return err
	}

	// Path: /nginx/v1/test
	// Methods: POST
	// Scopes: read, write
	// Description: Test nginx configuration
	if err := router.AddHandler(t, reTest, ctx.RequireScope(t.TestHandler, t.scoped, readScope, writeScope), http.MethodPost); err != nil {
		return err
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// HANDLERS

func (t *t) StatusHandler(w http.ResponseWriter, r *http.Request) {
	//	router.ResponseJSON(w, true, t.nginx.Version())
}

func (t *t) ReopenHandler(w http.ResponseWriter, r *http.Request) {
}

func (t *t) ReloadHandler(w http.ResponseWriter, r *http.Request) {
}

func (t *t) TestHandler(w http.ResponseWriter, r *http.Request) {
}
