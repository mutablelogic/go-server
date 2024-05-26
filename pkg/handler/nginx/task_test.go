package nginx_test

import (
	"context"
	"sync"
	"testing"
	"time"

	// Packages
	"github.com/mutablelogic/go-server/pkg/handler/nginx"
	"github.com/stretchr/testify/assert"
)

func Test_nginx_001(t *testing.T) {
	assert := assert.New(t)
	config := nginx.Config{}
	assert.NotEmpty(config.Name())
	assert.NotEmpty(config.Description())
}

func Test_nginx_002(t *testing.T) {
	assert := assert.New(t)
	task, err := nginx.New(nginx.Config{})
	assert.NoError(err)
	t.Log(task.Version())
}

func Test_nginx_003(t *testing.T) {
	var wg sync.WaitGroup

	// Create a new task
	assert := assert.New(t)
	task, err := nginx.New(nginx.Config{})
	assert.NoError(err)

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel and then wait for the task to finish
	defer wg.Wait()
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := task.Run(ctx); err != nil {
			t.Error(err)
		}
	}()

	// Sleep until the task is running
	time.Sleep(500 * time.Millisecond)
	err = task.Test()
	assert.NoError(err)
}
