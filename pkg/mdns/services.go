package mdns

import (
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/exp/slices"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Services stores the service names and registrations
type Services struct {
	sync.RWMutex
	discovered map[string]*Registration
	instances  map[string]*Service
}

type Registration struct {
	Browsed time.Time
	Expires time.Time
}

type Service struct {
	Name    string            `json:"name,omitempty"`
	Service string            `json:"service,omitempty"`
	Added   time.Time         `json:"added,omitempty"`
	Expires time.Time         `json:"expires,omitempty"`
	A       []net.IP          `json:"a,omitempty"`
	SRV     []Srv             `json:"srv,omitempty"`
	TXT     map[string]string `json:"txt,omitempty"`
	IfIndex int               `json:"-"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultServiceCap = 100
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// initialize the instance and registration maps
func (srv *Services) init() {
	if srv.discovered != nil && srv.instances != nil {
		return
	}
	srv.Lock()
	defer srv.Unlock()

	if srv.discovered == nil {
		srv.discovered = make(map[string]*Registration, defaultServiceCap)
	}
	if srv.instances == nil {
		srv.instances = make(map[string]*Service, defaultServiceCap)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get the list of service names which have not expired.
func (srv *Services) GetNames() []string {
	srv.init()
	srv.RLock()
	defer srv.RUnlock()

	result := make([]string, 0, len(srv.discovered))

	// Iterate over registrations, ignoring expired ones
	for k, v := range srv.discovered {
		if v.Expires.After(time.Now()) {
			result = append(result, k)
		}
	}

	// Iterate over services, ignoring expired ones
	for _, v := range srv.instances {
		if _, exists := srv.discovered[v.Service]; !exists && v.Expires.After(time.Now()) {
			result = append(result, v.Service)
		}
	}

	// Sort results and return
	sort.Strings(result)
	return result
}

// Returns a service name which should be browsed for instances. Will return an
// empty string if no service names are available, or if browsing has occurred within
// the ttl period.
func (srv *Services) NextBrowseName(ttl time.Duration) string {
	srv.init()
	srv.RLock()
	defer srv.RUnlock()

	// Make a list of non-expired service names
	keys := make([]string, 0, len(srv.discovered))
	for k, v := range srv.discovered {
		if v.Expires.After(time.Now()) {
			keys = append(keys, k)
		}
	}

	// Return if no service names
	if len(keys) == 0 {
		return ""
	}

	// Sort with oldest browsed first
	sort.Slice(keys, func(i, j int) bool {
		return srv.discovered[keys[i]].Browsed.Before(srv.discovered[keys[j]].Browsed)
	})

	// Return empty string if the first service has been browsed within the ttl period
	oldest := srv.discovered[keys[0]].Browsed
	if time.Since(oldest) < ttl {
		return ""
	}

	// Mark this service as browsed
	srv.discovered[keys[0]].Browsed = time.Now()

	// Retutn the first (oldest) service name
	return keys[0]
}

// Returns an instance name which should be looked up. Will return an empty	string
// if no instances are available
func (srv *Services) NextLookupInstance() *Service {
	srv.init()
	srv.RLock()
	defer srv.RUnlock()

	for _, v := range srv.instances {
		if v.Expires.After(time.Now()) {
			if len(v.SRV) == 0 {
				return v
			}
		}
	}

	// No instance found without resolved services
	return nil
}

// Get the list of service instances which have not expired.
func (srv *Services) GetInstances(services ...string) []*Service {
	srv.init()
	srv.RLock()
	defer srv.RUnlock()

	// Iterate over services, ignoring expired ones
	result := make([]*Service, 0, len(srv.instances))
	for _, v := range srv.instances {
		if v.Expires.After(time.Now()) {
			if len(services) == 0 || slices.Contains(services, v.Service) {
				result = append(result, v)
			}
		}
	}

	// Return result
	return result
}

// Expire service instances and return them
func (srv *Services) ExpireInstances() []*Service {
	srv.init()
	srv.Lock()
	defer srv.Unlock()

	// Iterate over services, adding expired ones to result
	result := make([]*Service, 0, len(srv.instances))
	for k, v := range srv.instances {
		if v.Expires.Before(time.Now()) {
			result = append(result, v)
			delete(srv.instances, k)
		}
	}

	// Return result
	return result
}

// Register a service name to the list of services, with a TTL
// and return true if the service was registered rather than updated
func (srv *Services) Registered(name string, ttl time.Duration) bool {
	srv.init()
	srv.Lock()
	defer srv.Unlock()
	_, exists := srv.discovered[name]
	if ttl > 0 {
		srv.discovered[name] = &Registration{Expires: time.Now().Add(ttl)}
	}
	return !exists
}

// AddPTR from service name to record, with a TTL, and return true if the
// service was registered rather than updated
func (srv *Services) AddPTR(key, service string, ptr Ptr, ifindex int) bool {
	srv.init()
	srv.Lock()
	defer srv.Unlock()
	if ptr.TTL() <= 0 {
		return false
	}

	// Add the instance record
	instance, exists := srv.instances[key]
	if !exists {
		instance = &Service{
			Added: time.Now(),
		}
	}
	instance.Expires = time.Now().Add(ptr.TTL())
	instance.IfIndex = ifindex
	instance.Name = strings.TrimSuffix(ptr.Name(), ptr.Service())
	instance.Service = service

	// Set the service and return
	srv.instances[key] = instance

	// Return true if the service was added
	return !exists
}

// AddTXT to instance
func (srv *Services) AddTXT(key string, txt []string) {
	srv.init()
	srv.Lock()
	defer srv.Unlock()

	// Add the instance record
	if instance, exists := srv.instances[key]; exists {
		instance.TXT = txtToMap(txt)
	} else {
		fmt.Println("NO KEY=", key)
	}
}

// AddSRV to instance
func (srv *Services) AddSRV(key string, r []Srv) {
	srv.init()
	srv.Lock()
	defer srv.Unlock()

	// Add the instance record
	if instance, exists := srv.instances[key]; exists {
		instance.SRV = r
	}
}

// AddAddr to instance
func (srv *Services) AddA(key string, r []net.IP) {
	srv.init()
	srv.Lock()
	defer srv.Unlock()

	// Add the instance record
	if instance, exists := srv.instances[key]; exists {
		instance.A = r
	}
}

// Expire a service instance and return true if the
// service was deleted
func (srv *Services) Expired(key string) bool {
	srv.init()
	if _, exists := srv.instances[key]; !exists {
		return false
	}
	srv.Lock()
	defer srv.Unlock()
	delete(srv.instances, key)
	return true
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func txtToMap(v []string) map[string]string {
	result := make(map[string]string, len(v))
	for _, s := range v {
		if s == "" {
			continue
		}
		kv := strings.SplitN(s, "=", 2)
		if len(kv) == 2 && kv[0] != "" {
			result[kv[0]] = kv[1]
		} else {
			result[s] = ""
		}
	}
	return result
}
