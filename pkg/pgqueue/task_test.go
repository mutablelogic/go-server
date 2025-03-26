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

func Test_Task(t *testing.T) {
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
	assert.Equal(types.PtrUint64(queue.Retries), uint64(10))

	t.Run("CreateTask", func(t *testing.T) {
		// Create a task
		task, err := client.CreateTask(context.TODO(), queue.Queue, schema.TaskMeta{
			Payload: "task_payload_1",
		})
		if assert.NoError(err) {
			assert.NotNil(task)
			assert.Equal(task.Queue, queue.Queue)
			assert.Equal(task.Payload, "task_payload_1")
			assert.Equal(types.PtrUint64(task.Retries), uint64(10))
			t.Log("Task created", task)
		}

		// Retain the task
		task2, err := client.RetainTask(context.TODO(), queue.Queue)
		if assert.NoError(err) {
			assert.NotNil(task2)
			assert.Equal(task2.Queue, queue.Queue)
			assert.Equal(task2.Payload, "task_payload_1")
			assert.Equal(types.PtrString(task2.Worker), client.Worker())
			assert.Equal(types.PtrUint64(task2.Retries), uint64(10))
			assert.NotNil(task2.CreatedAt)
			assert.NotNil(task2.StartedAt)
			assert.NotNil(task2.DiesAt)
			t.Log("Task retained", task2)
		}
	})

	t.Run("RetainTask", func(t *testing.T) {
		// Retain a task which does not exist
		task2, err := client.RetainTask(context.TODO(), queue.Queue)
		if assert.NoError(err) {
			assert.Nil(task2)
		}
	})

	t.Run("ReleaseTask", func(t *testing.T) {
		// Create a task
		task, err := client.CreateTask(context.TODO(), queue.Queue, schema.TaskMeta{
			Payload: "task_payload_2",
		})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(task)
		t.Log("Task created", task)

		// Retain the task
		task2, err := client.RetainTask(context.TODO(), queue.Queue)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(task2)
		assert.Equal(task2.Payload, "task_payload_2")
		t.Log("Task retained", task2)

		// Release the task
		task3, err := client.ReleaseTask(context.TODO(), types.PtrUint64(task2.Id), "completed")
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(task3)
		assert.Equal(task2.Payload, task3.Payload)
		assert.Equal("completed", task3.Result)
		assert.Equal(client.Worker(), types.PtrString(task3.Worker))
		assert.NotZero(types.PtrTime(task3.FinishedAt))
		t.Log("Task released", task3)
	})
}
