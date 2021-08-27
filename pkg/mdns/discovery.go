package mdns

import (
	"context"
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
	e        chan<- Event
	services map[string]*entry
}

type entry struct {
	service
	expires time.Time
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewDiscovery(c <-chan message, e chan<- Event) (*discovery, error) {
	this := new(discovery)
	this.c = c
	this.e = e
	this.services = make(map[string]*entry)

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
			} else if services := parsemessage(message.Msg, message.Zone); len(services) > 0 {
				for _, service := range services {
					this.process(service)
				}
			}
		}
	}
}

// Get returns a service by name, or nil if service not found
func (this *discovery) Get(name string) *service {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()

	key := unfqn(name)
	if entry, exists := this.services[key]; exists {
		return &entry.service
	} else {
		return nil
	}
}

// Check if service already exists
func (this *discovery) Exists(v *service) *service {
	return this.Get(v.Instance())
}

// Check if service has changed, returns false if
// service does not yet exist
func (this *discovery) Changed(v *service) bool {
	if other := this.Get(v.Instance()); other == nil {
		return false
	} else {
		return other.Equals(v)
	}
}

// Set a service and update expiration
func (this *discovery) Set(v *service) {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()

	key := unfqn(v.Instance())
	if v.ttl == 0 {
		// Delete service
		delete(this.services, key)
	} else {
		// Set service - grace period of 2 mins for expiry
		this.services[key] = &entry{
			*v,
			time.Now().Add(v.ttl).Add(2 * time.Minute),
		}
	}
}

// Delete a service
func (this *discovery) Delete(v *service) {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()

	// Delete service
	delete(this.services, unfqn(v.Instance()))
}

// Instances returns service instances
func (this *discovery) Instances() []*service {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()

	result := make([]*service, 0)
	for _, entry := range this.services {
		if time.Now().Before(entry.expires) {
			result = append(result, &entry.service)
		}
	}

	// Return services
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *discovery) send(e Event) {
	select {
	case this.e <- e:
		return
	default:
		panic("mDNS Discovery: Blocked channel")
	}
}

// Process a service record
func (this *discovery) process(service *service) {
	// Ignore if service is not valid
	if key := unfqn(service.Name()); key == "" {
		return
	}

	// Indicate service has changed
	if service.ttl == 0 {
		if this.Exists(service) != nil {
			this.send(Event{EVENT_TYPE_REMOVED, *service})
		}
	}

	// If this is a query, then indicate a service
	if unfqn(service.Service()) == ServicesQuery {
		this.send(Event{EVENT_TYPE_SERVICE, *service})
		return
	}

	// Check for added and changed services
	if other := this.Exists(service); other == nil {
		this.send(Event{EVENT_TYPE_ADDED, *service})
	} else if !other.Equals(service) {
		this.send(Event{EVENT_TYPE_CHANGED, *service})
	} else {
		return
	}

	// Add or remove the service
	this.Set(service)
}

// Expire service records
func (this *discovery) expire() {
	var expired []*service

	// Gather expired records
	this.RWMutex.Lock()
	for _, entry := range this.services {
		if entry.expires.Before(time.Now()) {
			expired = append(expired, &entry.service)
		}
	}
	this.RWMutex.Unlock()

	// Delete expired records
	for _, service := range expired {
		this.send(Event{EVENT_TYPE_EXPIRED, *service})
		this.Delete(service)
	}
}

// Parse DNS message and capture service records
func parsemessage(msg *dns.Msg, zone string) []*service {
	var result []*service
	sections := append(append(msg.Answer, msg.Ns...), msg.Extra...)
	for _, answer := range sections {
		switch rr := answer.(type) {
		case *dns.PTR:
			if len(result) == 0 {
				result = append(result, NewService(zone))
			}
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
