package pgqueue_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	// Packages
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
	assert "github.com/stretchr/testify/assert"
)

func Test_Task_1(t *testing.T) {
	assert := assert.New(t)
	conn := conn.Begin(t)
	defer conn.Close()

	// Create pgqueue
	client, err := pgqueue.New(context.TODO(), conn.PoolConn, pgqueue.OptWorker(t.Name()))
	assert.NoError(err)
	assert.NotNil(client)

	// Create queue
	queue, err := client.CreateQueue(context.TODO(), schema.Queue{
		Queue:      "queue_name_1",
		Retries:    types.Uint64Ptr(10),
		RetryDelay: types.DurationPtr(1 * time.Millisecond),
	})
	assert.NoError(err)
	assert.NotNil(queue)
	assert.Equal(types.PtrUint64(queue.Retries), uint64(10))

	t.Run("CreateTask", func(t *testing.T) {
		// Create a task
		task, err := client.CreateTask(context.TODO(), queue.Queue, schema.TaskMeta{
			Payload: "task_payload_1",
		})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(task)
		assert.Equal(task.Queue, queue.Queue)
		assert.Equal(task.Payload, "task_payload_1")
		assert.Equal(types.PtrUint64(task.Retries), uint64(10))
		t.Log("Task created", task)

		// Get the task
		var status string
		task1, err := client.GetTask(context.TODO(), task.Id, &status)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(task1)
		assert.Equal("new", status)
		assert.Equal(task1.Id, task.Id)

		// Retain the task
		task2, err := client.NextTask(context.TODO())
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

		// Get the task
		task3, err := client.GetTask(context.TODO(), task2.Id, &status)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(task3)
		assert.Equal("retained", status)
		assert.Equal(task2.Id, task.Id)

	})

	t.Run("RetainTask", func(t *testing.T) {
		// Retain a task which does not exist
		task2, err := client.NextTask(context.TODO())
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

		// Get the task
		var status string
		task1, err := client.GetTask(context.TODO(), task.Id, &status)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(task1)
		assert.Equal("new", status)
		assert.Equal(task1.Id, task.Id)

		// Retain the task
		task2, err := client.NextTask(context.TODO())
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(task2)
		assert.Equal(task2.Payload, "task_payload_2")
		t.Log("Task retained", task2)

		// Release the task
		task3, err := client.ReleaseTask(context.TODO(), task2.Id, "completed")
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

	t.Run("FailTask", func(t *testing.T) {
		// Create a task
		task, err := client.CreateTask(context.TODO(), queue.Queue, schema.TaskMeta{
			Payload: "task_payload_3",
		})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(task)
		t.Log("Task created", task)

		// Get the task
		var status string
		task1, err := client.GetTask(context.TODO(), task.Id, &status)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(task1)
		assert.Equal("new", status)
		assert.Equal(task1.Id, task.Id)

		retry := 1
		for {
			// Retain the task
			task2, err := client.NextTask(context.TODO())
			if !assert.NoError(err) {
				t.FailNow()
			}

			// No task ready yet - wait for some time
			if task2 == nil {
				time.Sleep(time.Duration(retry) * 100 * time.Millisecond)
				continue
			}

			assert.Equal(task2.Payload, "task_payload_3")
			t.Log("Task retained", task2)

			// Fail the task
			var status string
			task3, err := client.FailTask(context.TODO(), task2.Id, fmt.Sprint("failure_", retry), &status)
			if !assert.NoError(err) {
				t.FailNow()
			}
			assert.NotNil(task3)
			assert.True(status == "failed" || status == "retry")
			t.Log("Task failed", task3, " with status ", status)

			if status != "retry" {
				assert.Equal("failed", status)
				assert.Equal(types.PtrUint64(task3.Retries), uint64(0))
				break
			} else {
				assert.Equal("retry", status)
				assert.Greater(types.PtrUint64(task3.Retries), uint64(0))
				retry++
			}
		}
	})
}

func Test_Task_2(t *testing.T) {
	assert := assert.New(t)
	conn := conn.Begin(t)
	defer conn.Close()

	// Create pgqueue
	client, err := pgqueue.New(context.TODO(), conn.PoolConn, pgqueue.OptWorker(t.Name()))
	assert.NoError(err)
	assert.NotNil(client)

	// Create queue
	queue, err := client.CreateQueue(context.TODO(), schema.Queue{
		Queue:      "queue_name_2",
		Retries:    types.Uint64Ptr(10),
		RetryDelay: types.DurationPtr(1 * time.Millisecond),
	})
	assert.NoError(err)
	assert.NotNil(queue)
	assert.Equal(types.PtrUint64(queue.Retries), uint64(10))

	// Create a channel to receive tasks
	ch := make(chan *schema.Task)
	defer close(ch)

	// Create a channel to receive errors
	errch := make(chan error)
	defer close(errch)

	// Run the task loop and task insert in parallel
	var wg sync.WaitGroup

	// Create a context for the task loop
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := client.RunTaskLoop(ctx, ch, errch); err != nil {
			t.Log("Error running task loop:", err)
		}
		t.Log("Finished task loop")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range 10 {
			task, err := client.CreateTask(context.TODO(), queue.Queue, schema.TaskMeta{
				Payload: fmt.Sprintf("task_payload_%d", i),
			})
			if err != nil {
				t.Log("Error creating task:", err)
				errch <- err
				return
			}
			t.Log("Task created:", task)
			time.Sleep(time.Duration(500) * time.Millisecond)
		}
		t.Log("Finished emitting tasks")
	}()

	// Get tasks from the channel
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			t.Log("Context done, exiting loop")
			break FOR_LOOP
		case task := <-ch:
			t.Log("TASK:", task)
		case err := <-errch:
			t.Log("ERR:", err)
		}
	}

	// wait
	wg.Wait()
}
