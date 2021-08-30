package main

import (
	"context"
	"net/http"
	"regexp"
	"time"

	// Modules
	. "github.com/djthorpe/go-server"
	router "github.com/djthorpe/go-server/pkg/httprouter"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type EnumerateRequest struct {
	Timeout time.Duration `json:"timeout"`
	Service string        `json:"service"`
}

type ServiceResponse struct {
	Service     string `json:"service"`
	Description string `json:"description,omitempty"`
	Note        string `json:"note,omitempty"`
}

type InstanceResponse struct {
	Service  ServiceResponse   `json:"service"`
	Name     string            `json:"name"`
	Instance string            `json:"instance,omitempty"`
	Host     string            `json:"host,omitempty"`
	Port     uint16            `json:"port,omitempty"`
	Zone     string            `json:"zone,omitempty"`
	Addrs    []string          `json:"addrs,omitempty"`
	Txt      map[string]string `json:"txt,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// ROUTES

var (
	reRouteDatabase = regexp.MustCompile(`^/?$`)
)

///////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	minTimeout = time.Millisecond * 500
	maxTimeout = time.Second * 10
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (this *server) AddHandlers(ctx context.Context, provider Provider) error {
	// Add handler for returning instances in database
	if err := provider.AddHandlerFuncEx(ctx, reRouteDatabase, this.ServePing); err != nil {
		provider.Print(ctx, "Failed to add handler: ", err)
		return nil
	}

	// Add handler for enumerating services and instances, requires a timeout value to be posted
	if err := provider.AddHandlerFuncEx(ctx, reRouteDatabase, this.ServeEnumerate, http.MethodPost); err != nil {
		provider.Print(ctx, "Failed to add handler: ", err)
		return nil
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// HANDLERS

func (this *server) ServePing(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), time.Second)
	defer cancel()
	instances, err := this.EnumerateInstances(ctx)
	if err != nil {
		router.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	this.ServeInstances(w, instances)
}

func (this *server) ServeEnumerate(w http.ResponseWriter, req *http.Request) {
	// Parse request
	var request EnumerateRequest
	if err := router.RequestBody(req, &request); err != nil {
		router.ServeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Adjust timeout
	request.Timeout = minDuration(maxDuration(request.Timeout, minTimeout), maxTimeout)
	ctx, cancel := context.WithTimeout(req.Context(), request.Timeout)
	defer cancel()

	// Services are enumerated if no services in request
	if request.Service == "" {
		// Enumerate services
		services, err := this.EnumerateServices(ctx)
		if err != nil {
			router.ServeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Make response
		response := []ServiceResponse{}
		for _, service := range services {
			r := ServiceResponse{service, "", ""}
			if desc := this.LookupServiceDescription(service); desc != nil {
				r.Description = desc.Description()
				r.Note = desc.Note()
			}
			response = append(response, r)
		}

		// Serve response
		router.ServeJSON(w, response, http.StatusOK, 2)
	} else {
		// Enumerate instances
		instances, err := this.EnumerateInstances(ctx, request.Service)
		if err != nil {
			router.ServeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Make response
		this.ServeInstances(w, instances)
	}
}

func (this *server) ServeInstances(w http.ResponseWriter, instances []Service) {
	response := []InstanceResponse{}
	for _, instance := range instances {
		service := InstanceResponse{
			Service:  ServiceResponse{Service: instance.Service()},
			Zone:     instance.Zone(),
			Instance: instance.Instance(),
			Name:     instance.Name(),
			Host:     instance.Host(),
			Port:     instance.Port(),
			Addrs:    instanceAddrs(instance),
			Txt:      instanceTxt(instance),
		}
		if desc := this.LookupServiceDescription(instance.Service()); desc != nil {
			service.Service.Description = desc.Description()
			service.Service.Note = desc.Note()
		}
		response = append(response, service)
	}

	// Serve response
	router.ServeJSON(w, response, http.StatusOK, 2)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func instanceAddrs(instance Service) []string {
	addrs := instance.Addrs()
	result := make([]string, len(addrs))
	for i, addr := range addrs {
		result[i] = addr.String()
	}
	return result
}

func instanceTxt(instance Service) map[string]string {
	result := make(map[string]string)
	for _, key := range instance.Keys() {
		result[key] = instance.ValueForKey(key)
	}
	return result
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}
