package nginx

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	folders "github.com/mutablelogic/go-server/pkg/handler/nginx/folders"
	router "github.com/mutablelogic/go-server/pkg/handler/router"
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type responseHealth struct {
	Version string `json:"version"`
	Uptime  uint64 `json:"uptime"`
}

type responseTemplate struct {
	Name    string `json:"name,omitempty"`
	Enabled *bool  `json:"enabled,omitempty"` // Can be used for PATCH
	Body    string `json:"body,omitempty"`    // Can be used for PATCH
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	jsonIndent = 2
)

var (
	reRoot       = regexp.MustCompile(`^/?$`)
	reAction     = regexp.MustCompile(`^/(test|reload|reopen)/?$`)
	reListConfig = regexp.MustCompile(`^/config/?$`)
	reConfig     = regexp.MustCompile(`^/config/(.*)$`)
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - ENDPOINTS

// Add endpoints to the router
func (service *nginx) AddEndpoints(ctx context.Context, r server.Router) {
	// Path: /
	// Methods: GET
	// Scopes: read
	// Description: Get nginx status (version, uptime, available and enabled configurations)
	r.AddHandlerFuncRe(ctx, reRoot, service.GetHealth, http.MethodGet).(router.Route).
		SetScope(service.ScopeRead()...)

	// Path: /(test|reload|reopen)
	// Methods: PUT
	// Scopes: write
	// Description: Test, reload and reopen nginx configuration
	r.AddHandlerFuncRe(ctx, reAction, service.PutAction, http.MethodPut).(router.Route).
		SetScope(service.ScopeWrite()...)

	// Path: /config
	// Methods: GET
	// Scopes: read
	// Description: Read the current set of configurations
	r.AddHandlerFuncRe(ctx, reListConfig, service.ListConfig, http.MethodGet).(router.Route).
		SetScope(service.ScopeRead()...)

	// Path: /config
	// Methods: POST
	// Scopes: write
	// Description: Create a new configuration
	r.AddHandlerFuncRe(ctx, reListConfig, service.CreateConfig, http.MethodPost).(router.Route).
		SetScope(service.ScopeWrite()...)

	// Path: /config/{id}
	// Methods: GET
	// Scopes: read
	// Description: Read a configuration
	r.AddHandlerFuncRe(ctx, reConfig, service.ReadConfig, http.MethodGet).(router.Route).
		SetScope(service.ScopeRead()...)

	// Path: /config/{id}
	// Methods: DELETE, POST, PATCH
	// Scopes: write
	// Description: Modify a configuration
	r.AddHandlerFuncRe(ctx, reConfig, service.WriteConfig, http.MethodDelete, http.MethodPatch).(router.Route).
		SetScope(service.ScopeWrite()...)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get nginx status
func (service *nginx) GetHealth(w http.ResponseWriter, r *http.Request) {
	httpresponse.JSON(w, responseHealth{
		Version: string(service.Version()),
		Uptime:  uint64(time.Since(service.run.Start).Seconds()),
	}, http.StatusOK, jsonIndent)
}

// Do nginx action
func (service *nginx) PutAction(w http.ResponseWriter, r *http.Request) {
	urlParameters := router.Params(r.Context())
	switch urlParameters[0] {
	case "test":
		if err := service.Test(); err != nil {
			httpresponse.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
	case "reload":
		if err := service.Reload(); err != nil {
			httpresponse.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
	case "reopen":
		if err := service.Reopen(); err != nil {
			httpresponse.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
	default:
		httpresponse.Error(w, http.StatusNotFound)
		return
	}

	// Serve OK response with no body
	httpresponse.Empty(w, http.StatusOK)
}

// List configurations
func (service *nginx) ListConfig(w http.ResponseWriter, r *http.Request) {
	var result []responseTemplate
	for _, tmpl := range service.folders.Templates() {
		if response, err := service.tmplToResponse(tmpl); err != nil {
			httpresponse.Error(w, http.StatusInternalServerError, err.Error())
			return
		} else {
			result = append(result, *response)
		}
	}
	httpresponse.JSON(w, result, http.StatusOK, jsonIndent)
}

// Return a single configuration
func (service *nginx) ReadConfig(w http.ResponseWriter, r *http.Request) {
	urlParameters := router.Params(r.Context())
	tmpl := service.folders.Template(urlParameters[0])
	if tmpl == nil {
		httpresponse.Error(w, http.StatusNotFound)
		return
	} else if response, err := service.tmplToResponse(tmpl); err != nil {
		httpresponse.Error(w, http.StatusInternalServerError, err.Error())
		return
	} else {
		httpresponse.JSON(w, response, http.StatusOK, jsonIndent)
	}
}

// Create a new configuration
func (service *nginx) CreateConfig(w http.ResponseWriter, r *http.Request) {
	var create responseTemplate
	if err := httprequest.Read(r, &create); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, err.Error())
		return
	} else if create.Name == "" {
		httpresponse.Error(w, http.StatusBadRequest, "Missing name")
		return
	} else if create.Body == "" {
		httpresponse.Error(w, http.StatusBadRequest, "Missing body")
		return
	} else if tmpl := service.folders.Template(create.Name); tmpl != nil {
		httpresponse.Error(w, http.StatusConflict)
		return
	}

	// Create the configuration
	if err := service.folders.Create(create.Name, []byte(create.Body)); err != nil {
		httpresponse.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Check if we need to enable the configuration
	if create.Enabled != nil && *create.Enabled {
		if err := service.folders.Enable(create.Name); err != nil {
			httpresponse.Error(w, http.StatusInternalServerError, err.Error())
			return
		} else if err := service.Test(); err != nil {
			// Rollback
			err = errors.Join(err, service.folders.Delete(create.Name))
			httpresponse.Error(w, http.StatusInternalServerError, err.Error())
			return
		} else if err := service.Reload(); err != nil {
			httpresponse.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Respond with the template
	templ := service.folders.Template(create.Name)
	if templ == nil {
		httpresponse.Error(w, http.StatusNotFound, create.Name)
		return
	} else if response, err := service.tmplToResponse(templ); err != nil {
		httpresponse.Error(w, http.StatusInternalServerError, err.Error())
		return
	} else {
		httpresponse.JSON(w, response, http.StatusOK, jsonIndent)
	}
}

// PATCH or DELETE a configuration
func (service *nginx) WriteConfig(w http.ResponseWriter, r *http.Request) {
	urlParameters := router.Params(r.Context())
	templ := service.folders.Template(urlParameters[0])
	if templ == nil {
		httpresponse.Error(w, http.StatusNotFound)
		return
	}
	switch r.Method {
	case http.MethodDelete:
		service.DeleteConfig(templ, w, r)
	case http.MethodPatch:
		service.PatchConfig(templ, w, r)
	default:
		httpresponse.Error(w, http.StatusMethodNotAllowed, r.Method)
	}
}

// Delete a configuration
func (service *nginx) DeleteConfig(tmpl *folders.Template, w http.ResponseWriter, r *http.Request) {
	// Disable the configuration
	if tmpl.Enabled {
		if err := service.folders.Disable(tmpl.Name); err != nil {
			httpresponse.Error(w, http.StatusInternalServerError, err.Error())
			return
		} else if err := service.Reload(); err != nil {
			httpresponse.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Delete it
	if err := service.folders.Delete(tmpl.Name); err != nil {
		httpresponse.Error(w, http.StatusInternalServerError, err.Error())
	}

	// Respond with no content
	httpresponse.Empty(w, http.StatusOK)
}

// Update a configuration
func (service *nginx) PatchConfig(tmpl *folders.Template, w http.ResponseWriter, r *http.Request) {
	var patch responseTemplate
	var modified bool
	if err := httprequest.Read(r, &patch); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Body - currently only disabled configurations can have their body changed TODO: change this
	if patch.Body != "" {
		if patch.Enabled != nil && tmpl.Enabled {
			httpresponse.Error(w, http.StatusForbidden, "Cannot change body of enabled configuration")
			return
		}
		if err := service.folders.Write(tmpl.Name, []byte(patch.Body)); err != nil && errors.Is(err, ErrNotModified) {
			modified = false
		} else if err != nil {
			httpresponse.Error(w, http.StatusInternalServerError, err.Error())
			return
		} else {
			// Force update
			modified = true
			patch.Enabled = &tmpl.Enabled
		}
	}

	// Enabled
	if patch.Enabled != nil {
		modified = true
		switch *patch.Enabled {
		case true:
			if !tmpl.Enabled {
				if err := service.folders.Enable(tmpl.Name); err != nil {
					httpresponse.Error(w, http.StatusInternalServerError, err.Error())
					return
				}
			}
			if err := service.Test(); err != nil {
				// Rollback
				err = errors.Join(err, service.folders.Disable(tmpl.Name))
				httpresponse.Error(w, http.StatusInternalServerError, err.Error())
				return
			} else if err := service.Reload(); err != nil {
				httpresponse.Error(w, http.StatusInternalServerError, err.Error())
				return
			}
		case false:
			if tmpl.Enabled {
				if err := service.folders.Disable(tmpl.Name); err != nil {
					httpresponse.Error(w, http.StatusInternalServerError, err.Error())
					return
				}
			}
			if err := service.Reload(); err != nil {
				httpresponse.Error(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
	}

	// Return OK or Not Modified
	if modified {
		httpresponse.Empty(w, http.StatusOK)
	} else {
		httpresponse.Empty(w, http.StatusNotModified)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (service *nginx) tmplToResponse(tmpl *folders.Template) (*responseTemplate, error) {
	if body, err := service.folders.Render(tmpl.Name); err != nil {
		return nil, err
	} else {
		return &responseTemplate{
			Name:    tmpl.Name,
			Enabled: &tmpl.Enabled,
			Body:    string(body),
		}, nil
	}
}
