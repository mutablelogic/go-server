package mdns_test

import (
	"sync"
	"testing"
	"time"

	// Modules
	"github.com/mutablelogic/go-server/pkg/mdns"
	"golang.org/x/net/context"
)

func Test_Discovery_001(t *testing.T) {
	mdns, err := mdns.New(mdns.Config{})
	if err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := mdns.Run(ctx); err != nil {
			t.Error(err)
		}
	}()

	// Enumerate services
	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()
	if services, err := mdns.EnumerateServices(ctx2); err != nil {
		t.Error(err)
	} else {
		t.Log(services)
	}

	// Wait until done
	wg.Wait()
}

func Test_Discovery_002(t *testing.T) {
	mdns, err := mdns.New(mdns.Config{})
	if err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := mdns.Run(ctx); err != nil {
			t.Error(err)
		}
	}()

	// Enumerate services
	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()
	services, err := mdns.EnumerateServices(ctx2)
	if err != nil {
		t.Fatal(err)
	}
	ctx3, cancel3 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel3()
	instances, err := mdns.EnumerateInstances(ctx3, services...)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(instances)

	// Wait until done
	wg.Wait()
}
