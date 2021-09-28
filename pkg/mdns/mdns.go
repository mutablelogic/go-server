package mdns

import (
	"context"
	"fmt"
	"sync"
	"time"

	// Modules
	"github.com/hashicorp/go-multierror"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Interface       string   `yaml:"interface"`
	Domain          string   `yaml:"domain"`
	TTL             int      `yaml:"ttl"`
	ServiceDatabase []string `yaml:"services"`
}

type Server struct {
	*listener
	*discovery
	*servicedb
	c    chan message
	e    chan event
	C    chan Event // Channel for emitting events
	enum *enum
}

type enum struct {
	sync.RWMutex
	sync.Mutex
	EventType
	services map[string]*service
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ServicesQuery = "_services._dns-sd._udp"
	DefaultTTL    = 60 * 5 // In seconds (5 mins)
	defaultCap    = 100
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(cfg Config) (*Server, error) {
	this := new(Server)

	// Create channels and temp storage
	this.c = make(chan message, defaultCap)
	this.e = make(chan event, defaultCap)
	this.enum = &enum{}

	// Create listener
	if listener, err := NewListener(cfg.Domain, cfg.Interface, this.c); err != nil {
		return nil, err
	} else {
		this.listener = listener
	}

	// Create discovery to receive listener messages
	if discovery, err := NewDiscovery(this.c, this.e); err != nil {
		return nil, err
	} else {
		this.discovery = discovery
	}

	// Read service database
	if len(cfg.ServiceDatabase) > 0 {
		if db, err := ReadServiceDatabase(cfg.ServiceDatabase[0]); err != nil {
			return nil, err
		} else {
			this.servicedb = db
		}
		for _, source := range cfg.ServiceDatabase[1:] {
			if err := this.servicedb.Read(source); err != nil {
				return nil, err
			}
		}
	}

	// Return success
	return this, nil
}

func (this *Server) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	var result error

	// Run listener in background
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := this.listener.Run(ctx); err != nil {
			result = multierror.Append(result, err)
		}
	}()

	// Run discovery in background
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := this.discovery.Run(ctx); err != nil {
			result = multierror.Append(result, err)
		}
	}()

	// Run event processor in background
	wg.Add(1)
	go func() {
		defer wg.Done()
	FOR_LOOP:
		for {
			select {
			case event := <-this.e:
				// Add events to enumeration and also send to event bus
				this.send(event)
				this.enum.add(event)
			case <-ctx.Done():
				break FOR_LOOP
			}
		}
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	// Release resources
	close(this.c)
	close(this.e)

	// Return any errors
	return result
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Server) String() string {
	str := "<mdns"
	if this.listener != nil {
		str += " " + fmt.Sprint(this.listener)
	}
	if this.discovery != nil {
		str += " " + fmt.Sprint(this.discovery)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *Server) EnumerateServices(ctx context.Context) ([]string, error) {
	// Lock enumeration for receiving service
	this.enum.Lock(EVENT_TYPE_SERVICE, defaultCap)
	defer this.enum.Unlock()

	// Query for services on all interfaces
	query := msgQueryServices(this.listener.domain)
	if err := this.listener.Query(ctx, query, 0); err != nil {
		return nil, err
	}

	// Wait for context to complete
	<-ctx.Done()

	// Return service names
	var result []string
	for _, service := range this.enum.Services() {
		result = append(result, service.Instance())
	}
	return result, nil
}

// Return a service description for the given service name
func (this *Server) LookupServiceDescription(v string) ServiceDescription {
	if this.servicedb != nil {
		return this.servicedb.Lookup(v)
	} else {
		return nil
	}
}

func (this *Server) EnumerateInstances(ctx context.Context, services ...string) ([]Service, error) {
	// Lock enumeration for receiving instances
	this.enum.Lock(EVENT_TYPE_ADDED|EVENT_TYPE_CHANGED|EVENT_TYPE_REMOVED|EVENT_TYPE_EXPIRED, defaultCap)
	defer this.enum.Unlock()

	// Return if no services specified, then return instances from database
	if len(services) == 0 {
		instances := this.Instances()
		result := make([]Service, 0, len(instances))
		for _, instance := range instances {
			result = append(result, instance)
		}
		return result, nil
	}

	// Query for instances on all interfaces, cancel on context done
	ticker := time.NewTicker(emitRetryDuration * emitRetryCount)
	defer ticker.Stop()
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		case <-ticker.C:
			var service string
			service, services = services[0], services[1:]

			// Query for services on all interfaces
			query := msgQueryInstances(service, this.listener.domain)
			if err := this.listener.Query(ctx, query, 0); err != nil {
				return nil, err
			}

			// No more services to query
			if len(services) == 0 {
				break FOR_LOOP
			}
		}
	}

	// Wait for context to complete
	<-ctx.Done()

	// Return services
	result := make([]Service, 0, len(this.enum.services))
	for _, instance := range this.enum.services {
		result = append(result, instance)
	}

	return result, nil
}

// Lock enumeration
func (this *enum) Lock(t EventType, cap int) {
	this.RWMutex.Lock()
	this.EventType = t
	this.services = make(map[string]*service, cap)
}

// Unlock enumeration
func (this *enum) Unlock() {
	this.EventType = EVENT_TYPE_NONE
	this.services = nil
	this.RWMutex.Unlock()
}

// Services returns enumerated services
func (this *enum) Services() []*service {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	result := make([]*service, 0, len(this.services))
	for _, service := range this.services {
		// Make a copy of the service
		result = append(result, service)
	}
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// add filters an added service in the enumeration
func (this *enum) add(event event) {
	if this.EventType == EVENT_TYPE_NONE || event.EventType&this.EventType != EVENT_TYPE_NONE {
		// Add or remove filtered event, keyed by service name
		this.Mutex.Lock()
		defer this.Mutex.Unlock()
		if this.services != nil {
			key := event.Instance()
			if event.EventType&(EVENT_TYPE_REMOVED|EVENT_TYPE_EXPIRED) != EVENT_TYPE_NONE {
				delete(this.services, key)
			} else {
				this.services[key] = &event.service
			}
		}
	}
}

// send messages on the channel
func (this *Server) send(event event) {
	select {
	case this.C <- event:
		return
	default:
		return
	}
}
