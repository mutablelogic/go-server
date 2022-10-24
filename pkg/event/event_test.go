package event_test

import (
	"context"
	"errors"
	"testing"

	// Packages
	event "github.com/mutablelogic/go-server/pkg/event"
)

func Test_Event_000(t *testing.T) {
	t.Log(t.Name())
}

func Test_Event_001(t *testing.T) {
	e := event.New(context.TODO(), "key", "value")
	if e == nil {
		t.Fatal("Expected non-nil")
	}
	if e.Key() != "key" {
		t.Fatal("Expected key")
	}
	if e.Value() != "value" {
		t.Fatal("Expected value")
	}
}

func Test_Event_002(t *testing.T) {
	e := event.New(context.TODO(), nil, "value")
	if e != nil {
		t.Fatal("Expected nil")
	}
}

func Test_Event_003(t *testing.T) {
	err := errors.New("Error")
	e := event.Error(context.TODO(), err)
	if e == nil {
		t.Fatal("Expected non-nil")
	}
	if e.Key() != nil {
		t.Fatal("Expected nil key")
	}
	if !errors.Is(e.Value().(error), err) {
		t.Fatal("Unexpected error return")
	}
	if !errors.Is(e.Error(), err) {
		t.Fatal("Unexpected error return")
	}
}
