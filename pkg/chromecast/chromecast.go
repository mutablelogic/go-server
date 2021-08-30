package chromecast

import (
	"context"
	"fmt"
	"sync"
	"time"

	// Modules
	. "github.com/djthorpe/go-server"
	multierror "github.com/hashicorp/go-multierror"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
}

type Manager struct {
	sync.RWMutex
	cast map[string]*Cast
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ServiceName           = "_googlecast._tcp."
	serviceConnectTimeout = time.Second * 15
	serciceMessageTimeout = time.Second
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewManager(cfg Config) *Manager {
	this := new(Manager)
	this.cast = make(map[string]*Cast)

	// Return success
	return this
}

func (this *Manager) Run(ctx context.Context) error {
	// TODO: Update status

	// Wait for end of run
	<-ctx.Done()

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
// PUBLIC METHODS

func (this *Manager) Add(instance Service) error {
	key := instance.Instance()
	if cast := this.Get(key); cast != nil {
		fmt.Println("chromecast update", cast)
	} else {
		this.RWMutex.Lock()
		defer this.RWMutex.Unlock()
		this.cast[key] = NewCast(instance)
		fmt.Println("chromecast added", this.cast[key])
	}

	// Return success
	return nil
}

func (this *Manager) Remove(instance Service) error {
	key := instance.Instance()
	if cast := this.Get(key); cast != nil {
		fmt.Println("chromecast deleted", cast)
		this.RWMutex.Lock()
		defer this.RWMutex.Unlock()
		delete(this.cast, key)
	}

	// Return success
	return nil
}

func (this *Manager) Get(id string) *Cast {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()
	if cast, exists := this.cast[id]; !exists {
		return nil
	} else {
		return cast
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Manager) String() string {
	str := "<chromecast"
	for _, cast := range this.cast {
		str += fmt.Sprint(" ", cast)
	}
	return str + ">"
}
