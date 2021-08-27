package chromecast

import (
	"context"
	"fmt"
	"sync"
	"time"

	// Modules

	multierror "github.com/hashicorp/go-multierror"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ServiceDiscovery interface {
	EnumerateInstances(ctx context.Context, serviceType string) ([]Service, error)
}

type Config struct {
	ServiceDiscovery
}

type chromecast struct {
	sync.RWMutex
	ServiceDiscovery
	cast map[string]*device
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	serviceTypeCast       = "_googlecast._tcp."
	serviceConnectTimeout = time.Second * 15
	serciceMessageTimeout = time.Second
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(cfg Config) error {
	this := new(chromecast)
	this.cast = make(map[string]*device)

	// Set service discovery
	this.ServiceDiscovery = cfg.ServiceDiscovery

	// Return success
	return nil
}

func (this *chromecast) Run(ctx context.Context) error {
	// Disconnect devices
	var result error
	for _, cast := range this.cast {
		// TODO: remove cast
		if err := cast.disconnect(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Release resources
	this.cast = nil

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *chromecast) String() string {
	str := "<chromecast"
	for _, cast := range this.cast {
		str += fmt.Sprint(" ", cast)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *chromecast) devices(ctx context.Context) ([]*device, error) {
	// Perform the lookup
	records, err := this.ServiceDiscovery.EnumerateInstances(ctx, serviceTypeCast)
	if err != nil {
		return nil, err
	}

	// Return any casts found
	result := make([]*device, 0, len(records))
	for _, record := range records {
		/*
			cast := NewCastFromRecord(record)
			if cast == nil {
				continue
			}

			// Add cast, emit event
			if existing := this.getCastForId(cast.id); existing == nil {
				this.castevent(cast)
			}

			// Append cast onto results
			result = append(result, this.getCastForId(cast.id))
		*/
	}

	// Return success
	return result, nil
}
