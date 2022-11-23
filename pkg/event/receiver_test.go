package event_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	// Packages
	event "github.com/mutablelogic/go-server/pkg/event"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

func Test_Receiver_000(t *testing.T) {
	var rcv event.Receiver
	t.Log(&rcv)
}

func Test_Receiver_001(t *testing.T) {
	var src event.Source
	var rcv event.Receiver
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	evt := event.New(context.TODO(), "key", "value")
	got := false

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := rcv.Rcv(ctx, func(e Event) error {
			t.Log("Processing message", e)
			if e != evt {
				t.Error("Expected evt")
			} else {
				got = true
			}
			return nil
		}, &src); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			t.Error(err)
		}
	}()

	if ok := src.Emit(evt); !ok {
		t.Error("Got false from emit function")
	} else {
		t.Log("Sent event", evt)
	}

	// Wait for timeout
	wg.Wait()

	// Check received messages
	if !got {
		t.Error("Expected event, but it wasn't emitted")
	}
}
