package mdns

import (
	"context"
	"fmt"
	"sync"
	"time"

	// Modules
	dns "github.com/miekg/dns"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type discovery struct {
	sync.RWMutex

	c        <-chan message
	services map[string]entry
}

type entry struct {
	Service
	expires time.Time
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewDiscovery(c <-chan message) (*discovery, error) {
	this := new(discovery)
	this.c = c
	this.services = make(map[string]entry)

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Run discovery in background until cancelled
func (this *discovery) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			this.expire()
		case message := <-this.c:
			if message.Err != nil {
				// An error occurred, ignore for now
				// fmt.Println("error=", message.Err)
			} else if services := parsemessage(message.Msg, message.Zone); len(services) > 0 {
				for _, service := range services {
					this.process(service)
				}
			}
		}
	}
}

// Get returns a service by name, or nil if service not found
func (this *discovery) Get(name string) *Service {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()

	key := unfqn(name)
	if entry, exists := this.services[key]; exists {
		return &entry.Service
	} else {
		return nil
	}
}

// Check if service already exists
func (this *discovery) Exists(v *Service) *Service {
	return this.Get(v.Instance())
}

// Check if service has changed, returns false if
// service does not yet exist
func (this *discovery) Changed(v *Service) bool {
	if other := this.Get(v.Instance()); other == nil {
		return false
	} else {
		return other.Equals(v)
	}
}

// Set a service and update expiration
func (this *discovery) Set(v *Service) {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()

	key := unfqn(v.Instance())
	if v.ttl == 0 {
		// Delete service
		delete(this.services, key)
	} else {
		// Set service
		this.services[key] = entry{
			*v,
			time.Now().Add(v.ttl).Add(time.Minute),
		}
	}
}

// Delete a service
func (this *discovery) Delete(v *Service) {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()

	// Delete service
	delete(this.services, unfqn(v.Instance()))
}

// Instances returns service instances for a service name, or all services
func (this *discovery) Instances(name string) []Service {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()

	result := make([]Service, 0)
	key := fqn(name)
	for _, entry := range this.services {
		if time.Now().Before(entry.expires) && (name == "" || entry.Service.Service() == key) {
			result = append(result, entry.Service)
		}
	}

	// Return services
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Process a service record
func (this *discovery) process(service *Service) {
	// Ignore if service is not valid
	if key := unfqn(service.Name()); key == "" {
		return
	} else if service.ttl == 0 {
		if this.Exists(service) != nil {
			fmt.Println("REMOVED", service)
		}
	} else if other := this.Exists(service); other == nil {
		fmt.Println("ADDED", service)
	} else if !other.Equals(service) {
		fmt.Println("CHANGED", service)
	}
	this.Set(service)
}

// Expire service records
func (this *discovery) expire() {
	var expired []*Service

	// Gather expired records
	this.RWMutex.Lock()
	for _, entry := range this.services {
		if entry.expires.Before(time.Now()) {
			fmt.Println("EXPIRED", entry.Name(), entry.expires)
			expired = append(expired, &entry.Service)
		}
	}
	this.RWMutex.Unlock()

	// Delete expired records
	for _, service := range expired {
		fmt.Println("EXPIRED", service)
		this.Delete(service)
	}
}

// Parse DNS message and capture service records
func parsemessage(msg *dns.Msg, zone string) []*Service {
	var result []*Service
	sections := append(append(msg.Answer, msg.Ns...), msg.Extra...)
	for _, answer := range sections {
		switch rr := answer.(type) {
		case *dns.PTR:
			result = append(result, NewService(zone))
			result[0].SetPTR(rr)
		case *dns.SRV:
			if len(result) > 0 {
				result[0].SetSRV(rr.Target, rr.Port, rr.Priority)
			}
		case *dns.TXT:
			if len(result) > 0 {
				result[0].SetTXT(rr.Txt)
			}
		case *dns.A:
			if len(result) > 0 {
				result[0].SetA(rr.A)
			}
		case *dns.AAAA:
			if len(result) > 0 {
				result[0].SetAAAA(rr.AAAA)
			}
		}
	}

	// Return any services
	return result
}
