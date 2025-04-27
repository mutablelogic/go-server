package pgqueue_test

import (
	"context"
	"testing"
	"time"

	// Packages
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	assert "github.com/stretchr/testify/assert"
)

/////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Test_TaskPool_001(t *testing.T) {
	assert := assert.New(t)
	pool := pgqueue.NewTaskPool(1)
	defer pool.Close()
	assert.NotNil(pool)
}

func Test_TaskPool_002(t *testing.T) {
	assert := assert.New(t)
	pool := pgqueue.NewTaskPool(5)
	assert.NotNil(pool)

	j := 0
	for i := 0; i < 10; i++ {
		pool.RunTask(context.TODO(), &schema.Task{}, func(ctx context.Context, payload any) error {
			t.Log("Task with payload", payload)
			time.Sleep(time.Second)
			j += 1
			return nil
		}, i, nil)
	}
	pool.Close()
	assert.Equal(10, j)
}
