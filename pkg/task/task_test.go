package task_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	// Package imports
	"github.com/mutablelogic/go-server/pkg/task"
)

func Test_Task_000(t *testing.T) {
	tsk := new(task.Task)

	var err error
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = tsk.Run(ctx)
	}()
	cancel()
	wg.Wait()
	if !errors.Is(err, context.Canceled) {
		t.Error("Expected context.Canceled, got", err)
	}
}
