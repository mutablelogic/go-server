package mdns

import (
	"context"
	"fmt"
	"sync"

	// Modules
	"github.com/hashicorp/go-multierror"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Interface       string `yaml:"interface"`
	Domain          string `yaml:"domain"`
	TTL             int    `yaml:"ttl"`
	ServiceDatabase string `yaml:"services"`
}

type Server struct {
	*listener
	*discovery
	*servicedb
	c chan message
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ServicesQuery = "_services._dns-sd._udp"
	DefaultTTL    = 60 * 5 // In seconds (5 mins)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(cfg Config) (*Server, error) {
	this := new(Server)

	// Create channel
	this.c = make(chan message)

	// Create listener
	if listener, err := NewListener(cfg.Domain, cfg.Interface, this.c); err != nil {
		return nil, err
	} else {
		this.listener = listener
	}

	// Create discovery to receive listener messages
	if discovery, err := NewDiscovery(this.c); err != nil {
		return nil, err
	} else {
		this.discovery = discovery
	}

	// Read service database in the background
	if cfg.ServiceDatabase != "" {
		go func() {
			if db, err := ReadServiceDatabase(cfg.ServiceDatabase); err != nil {
				fmt.Printf("Failed to read service database: %s\n", err)
			} else {
				this.servicedb = db
			}
		}()
	}

	// Return success
	return this, nil
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

	// Wait for all goroutines to complete
	wg.Wait()

	// Release resources
	close(this.c)

	// Return any errors
	return result
}

/*
func (this *Server) EnumerateServices(ctx context.Context) ([]string, error) {
	// Query for services on all interfaces
	query := msgQueryServices(this.listener.domain)
	if err := this.listener.Query(ctx, query, 0); err != nil {
		return nil, err
	}

	<-ctx.Done()

	// TODO: Collect names
	return []string{}, nil
}
*/

func (this *Server) Instances(ctx context.Context, services ...string) []Service {
	// TODO: Query for services

	// Shortcut for all services
	if len(services) == 0 {
		return this.discovery.Instances("")
	}

	// Filter through instances
	instances := make(map[string]Service)
	for _, service := range services {
		if service == "" {
			continue
		}
		for _, instance := range this.discovery.Instances(service) {
			if _, exists := instances[instance.Instance()]; !exists {
				instances[instance.Instance()] = instance
			}
		}
	}

	// Return instance array
	result := make([]Service, len(instances))
	for _, instance := range instances {
		result = append(result, instance)
	}
	return result
}
