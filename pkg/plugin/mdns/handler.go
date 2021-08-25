package main

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"

	// Modules
	router "github.com/djthorpe/go-server/pkg/httprouter"
	mdns "github.com/djthorpe/go-server/pkg/mdns"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Service struct {
	Service string            `json:"service"`
	Zone    string            `json:"zone"`
	Name    string            `json:"name"`
	Host    string            `json:"host,omitempty"`
	Port    uint16            `json:"port,omitempty"`
	Addrs   []string          `json:"addrs,omitempty"`
	Txt     map[string]string `json:"txt,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// ROUTES

var (
	reRouteInstances = regexp.MustCompile(`^/?$`)
)

///////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	maxResultLimit = 1000
)

///////////////////////////////////////////////////////////////////////////////
// HANDLERS

func (this *server) ServeInstances(w http.ResponseWriter, req *http.Request) {
	// Get all instances
	ctx, cancel := context.WithTimeout(req.Context(), time.Millisecond*50)
	defer cancel()
	instances := this.Instances(ctx)

	// Make response
	response := []Service{}
	for _, instance := range instances {
		response = append(response, Service{
			Service: instance.Service(),
			Zone:    instance.Zone(),
			Name:    instance.Name(),
			Host:    instance.Host(),
			Port:    instance.Port(),
			Addrs:   instanceAddrs(&instance),
			Txt:     instanceTxt(&instance),
		})
	}

	// Serve response
	router.ServeJSON(w, response, http.StatusOK, 2)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func instanceAddrs(instance *mdns.Service) []string {
	addrs := instance.Addrs()
	result := make([]string, len(addrs))
	for i, addr := range addrs {
		result[i] = addr.String()
	}
	return result
}

func instanceTxt(instance *mdns.Service) map[string]string {
	txt := instance.Txt()
	result := make(map[string]string, len(txt))
	for _, value := range txt {
		if kv := strings.SplitN(value, "=", 2); len(kv) == 2 {
			result[kv[0]] = kv[1]
		} else {
			result[value] = ""
		}
	}
	return result
}
