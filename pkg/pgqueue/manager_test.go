package pgqueue_test

import (
	"context"
	"sync"
	"testing"
	"time"

	// Packages
	test "github.com/djthorpe/go-pg/pkg/test"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
	assert "github.com/stretchr/testify/assert"
)

// Global connection variable
var conn test.Conn

// Start up a container and test the pool
func TestMain(m *testing.M) {
	test.Main(m, &conn)
}

/////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Test_Manager_001(t *testing.T) {
	assert := assert.New(t)
	conn := conn.Begin(t)
	defer conn.Close()

	// Create a new queue manager
	manager, err := pgqueue.NewManager(context.TODO(), conn)
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(manager)

	// Register a ticker
	t.Run("RegisterTicker1", func(t *testing.T) {
		ticker, err := manager.RegisterTicker(context.TODO(), schema.TickerMeta{
			Ticker: "ticker1",
		})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(ticker)
		assert.Equal("ticker1", ticker.Ticker)
		assert.NotNil(ticker.Interval)
		assert.Equal(60*time.Second, *ticker.Interval)
		assert.Nil(ticker.Ts)
	})

	// Re-register a ticker with a different interval
	t.Run("RegisterTicker2", func(t *testing.T) {
		ticker1, err := manager.RegisterTicker(context.TODO(), schema.TickerMeta{
			Ticker: "ticker2",
		})
		if !assert.NoError(err) {
			t.FailNow()
		}

		ticker2, err := manager.RegisterTicker(context.TODO(), schema.TickerMeta{
			Ticker:   "ticker2",
			Interval: types.DurationPtr(120 * time.Second),
		})
		if !assert.NoError(err) {
			t.FailNow()
		}

		assert.Equal("ticker2", ticker1.Ticker)
		assert.Equal("ticker2", ticker2.Ticker)
		assert.Equal(60*time.Second, *ticker1.Interval)
		assert.Equal(120*time.Second, *ticker2.Interval)
	})

	// Delete a ticker
	t.Run("DeleteTicker1", func(t *testing.T) {
		_, err := manager.DeleteTicker(context.TODO(), "nonexistent")
		assert.ErrorIs(err, httpresponse.ErrNotFound)
	})

	// Delete a ticker
	t.Run("DeleteTicker2", func(t *testing.T) {
		ticker1, err := manager.RegisterTicker(context.TODO(), schema.TickerMeta{
			Ticker: "ticker3",
		})
		if !assert.NoError(err) {
			t.FailNow()
		}

		ticker2, err := manager.DeleteTicker(context.TODO(), "ticker3")
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(ticker1, ticker2)
	})

	// List tickers
	t.Run("ListTickers", func(t *testing.T) {
		list, err := manager.ListTickers(context.TODO(), schema.TickerListRequest{})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(list)
		assert.NotZero(list.Count)
		assert.NotNil(list.Body)
	})

	// Next ticker
	t.Run("NextTicker", func(t *testing.T) {
		_, err := manager.RegisterTicker(context.TODO(), schema.TickerMeta{
			Ticker: "ticker4",
		})
		if !assert.NoError(err) {
			t.FailNow()
		}

		var count uint64
		for {
			ticker, err := manager.NextTicker(context.TODO())
			if !assert.NoError(err) {
				t.FailNow()
			}
			if ticker == nil {
				break
			}
			count++
		}
		assert.NotZero(count)
	})

	// Loop ticker
	t.Run("LoopTicker", func(t *testing.T) {
		_, err := manager.RegisterTicker(context.TODO(), schema.TickerMeta{
			Ticker:   "ticker5",
			Interval: types.DurationPtr(1 * time.Second),
		})
		if !assert.NoError(err) {
			t.FailNow()
		}

		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()

		ch := make(chan *schema.Ticker, 10)
		defer close(ch)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := manager.RunTickerLoop(ctx, ch); err != nil {
				assert.NoError(err)
			}
		}()
		for {
			select {
			case <-ctx.Done():
				wg.Wait()
				return
			case ticker := <-ch:
				assert.NotNil(ticker)
				t.Log(ticker)
			}
		}
	})
}

func Test_Manager_002(t *testing.T) {
	assert := assert.New(t)
	conn := conn.Begin(t)
	defer conn.Close()

	// Create a new queue manager
	manager, err := pgqueue.NewManager(context.TODO(), conn)
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(manager)

	// Register a queue
	t.Run("RegisterQueue1", func(t *testing.T) {
		queue, err := manager.RegisterQueue(context.TODO(), schema.Queue{
			Queue: "queue1",
		})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(queue)
		assert.Equal("queue1", queue.Queue)
		assert.Equal(time.Hour, *queue.TTL)
		assert.Equal(uint64(3), *queue.Retries)
		assert.Equal(2*time.Minute, *queue.RetryDelay)
	})

	// Re-register a queue with different parameters
	t.Run("RegisterTicker2", func(t *testing.T) {
		queue1, err := manager.RegisterQueue(context.TODO(), schema.Queue{
			Queue: "queue2",
		})
		if !assert.NoError(err) {
			t.FailNow()
		}
		queue2, err := manager.RegisterQueue(context.TODO(), schema.Queue{
			Queue:      "queue2",
			TTL:        types.DurationPtr(2 * time.Hour),
			Retries:    types.Uint64Ptr(5),
			RetryDelay: types.DurationPtr(10 * time.Minute),
		})
		if !assert.NoError(err) {
			t.FailNow()
		}

		assert.Equal("queue2", queue1.Queue)
		assert.Equal("queue2", queue2.Queue)
		assert.Equal(2*time.Hour, *queue2.TTL)
		assert.Equal(uint64(5), *queue2.Retries)
		assert.Equal(10*time.Minute, *queue2.RetryDelay)
	})

	// Get non-existent queue
	t.Run("GetQueue1", func(t *testing.T) {
		queue, err := manager.GetQueue(context.TODO(), "nonexistent")
		assert.ErrorIs(err, httpresponse.ErrNotFound)
		assert.Nil(queue)
	})

	// Get queue
	t.Run("GetQueue2", func(t *testing.T) {
		queue1, err := manager.RegisterQueue(context.TODO(), schema.Queue{
			Queue: "queue3",
		})
		if !assert.NoError(err) {
			t.FailNow()
		}

		queue2, err := manager.GetQueue(context.TODO(), "queue3")
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(queue1, queue2)
	})

	// List queues
	t.Run("ListQueues", func(t *testing.T) {
		list, err := manager.ListQueues(context.TODO(), schema.QueueListRequest{})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(list)
		assert.NotZero(list.Count)
		assert.NotNil(list.Body)
	})
}
