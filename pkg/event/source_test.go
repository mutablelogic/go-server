package event_test

import (
	"context"
	"sync"
	"testing"

	// Packages
	event "github.com/mutablelogic/go-server/pkg/event"
)

func Test_Source_000(t *testing.T) {
	var src event.Source
	t.Log(&src)
}

func Test_Source_001(t *testing.T) {
	var src event.Source
	if ok := src.Emit(event.New(context.TODO(), "key", "value")); !ok {
		t.Fatal("Expected ok")
	}
}

func Test_Source_002(t *testing.T) {
	var src event.Source
	ch := src.Sub()
	t.Log(&src)
	src.Unsub(ch)
	t.Log(&src)
}

func Test_Source_003(t *testing.T) {
	var src event.Source
	var wg sync.WaitGroup

	evt := event.New(context.TODO(), "key", "value")
	ch := src.Sub()
	wg.Add(1)
	go func() {
		defer wg.Done()
		e := <-ch
		if e != evt {
			t.Error("Expected", evt, " got ", e)
		}
	}()
	if ok := src.Emit(evt); !ok {
		t.Fatal("Got false from emit")
	}
	wg.Wait()
	src.Unsub(ch)
}
