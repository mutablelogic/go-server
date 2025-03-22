package pgqueue_test

import (
	"context"
	"testing"
	"time"

	// Packages
	test "github.com/djthorpe/go-pg/pkg/test"
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue"
	"github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	"github.com/mutablelogic/go-service/pkg/types"
	assert "github.com/stretchr/testify/assert"
)

// Global connection variable
var conn test.Conn

// Start up a container and test the pool
func TestMain(m *testing.M) {
	test.Main(m, &conn)
}

func Test_Queue_001(t *testing.T) {
	assert := assert.New(t)
	conn := conn.Begin(t)
	defer conn.Close()

	// Ping the database
	assert.NoError(conn.Ping(context.Background()))

	// Create pgqueue
	client, err := pgqueue.New(context.TODO(), conn.PoolConn, pgqueue.OptWorker(t.Name()))
	assert.NoError(err)
	assert.NotNil(client)

	// Create queue
	t.Run("CreateQueue_1", func(t *testing.T) {
		queue, err := client.CreateQueue(context.TODO(), schema.Queue{
			Queue: "queue_name_1",
		})
		assert.NoError(err)
		assert.NotNil(queue)
	})

	// Create queue
	t.Run("CreateQueue_2", func(t *testing.T) {
		queue, err := client.CreateQueue(context.TODO(), schema.Queue{
			Queue:      "queue_name_2",
			TTL:        types.DurationPtr(5 * time.Hour),
			Retries:    types.Uint64Ptr(10),
			RetryDelay: types.DurationPtr(5 * time.Minute),
		})
		assert.NoError(err)
		assert.NotNil(queue)
	})

	// Create queue then get queue
	t.Run("GetQueue", func(t *testing.T) {
		queue, err := client.CreateQueue(context.TODO(), schema.Queue{
			Queue:      "queue_name_3",
			TTL:        types.DurationPtr(10 * time.Hour),
			Retries:    types.Uint64Ptr(10),
			RetryDelay: types.DurationPtr(5 * time.Minute),
		})
		assert.NoError(err)
		assert.NotNil(queue)

		queue2, err := client.GetQueue(context.TODO(), queue.Queue)
		assert.NoError(err)
		assert.NotNil(queue2)
		assert.Equal(queue.Queue, queue2.Queue)
	})

	// Get queue
	t.Run("GetQueue_2", func(t *testing.T) {
		queue, err := client.GetQueue(context.TODO(), "non_extistent_queue")
		assert.NoError(err)
		assert.NotNil(queue)
		t.Log(queue)
	})
	/*
	   // Create queue then delete queue

	   	t.Run("DeleteQueue", func(t *testing.T) {
	   		queue, err := client.CreateQueue(context.TODO(), schema.Queue{
	   			Queue: "queue_name_4",
	   		})
	   		assert.NoError(err)
	   		assert.NotNil(queue)

	   		queue2, err := client.DeleteQueue(context.TODO(), "q")
	   		assert.NoError(err)
	   		assert.NotNil(queue2)
	   		assert.Equal(queue.Queue, queue2.Queue)

	   		_, err = client.GetQueue(context.TODO(), queue.Queue)
	   		assert.NoError(err)
	   	})
	*/
}
