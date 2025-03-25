package pgqueue_test

import (
	"context"
	"testing"

	// Packages
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
	assert "github.com/stretchr/testify/assert"
)

func Test_Task_001(t *testing.T) {
	assert := assert.New(t)
	conn := conn.Begin(t)
	defer conn.Close()

	// Create pgqueue
	client, err := pgqueue.New(context.TODO(), conn.PoolConn, pgqueue.OptWorker(t.Name()))
	assert.NoError(err)
	assert.NotNil(client)

	// Create queue
	queue, err := client.CreateQueue(context.TODO(), schema.Queue{
		Queue:   "queue_name_1",
		Retries: types.Uint64Ptr(10),
	})
	assert.NoError(err)
	assert.NotNil(queue)
	assert.Equal(types.PtrUint64(queue.Retries), 10)

	// Create a task
	t.Run("CreateTask_1", func(t *testing.T) {
		task, err := client.CreateTask(context.TODO(), queue.Queue, schema.TaskMeta{
			Payload: "task_payload_1",
		})
		if assert.NoError(err) {
			assert.NotNil(task)
			assert.Equal(task.Queue, queue.Queue)
			assert.Equal(task.Payload, "task_payload_1")
			assert.Equal(types.PtrUint64(task.Retries), 10)
			t.Log("Task created", task)
		}
	})
}
