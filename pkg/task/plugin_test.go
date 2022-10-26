package task_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	// Package imports
	"github.com/mutablelogic/go-server/pkg/task"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

func Test_Plugin_000(t *testing.T) {
	var plugin = task.Plugin{}
	if err := json.Unmarshal([]byte(`{}`), &plugin); err != nil {
		t.Error("Unexpected error", err)
	} else if _, err := plugin.New(context.TODO(), nil); !errors.Is(err, ErrBadParameter) {
		t.Error("Expected ErrBadParameter, got", err)
	}
}

func Test_Plugin_001(t *testing.T) {
	var plugins = task.Plugins{}
	if err := plugins.Register(task.Plugin{"name", "label"}); err != nil {
		t.Error("Unexpected error", err)
	}
}
