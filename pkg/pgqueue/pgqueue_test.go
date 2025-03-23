package pgqueue_test

import (
	"context"
	"testing"
	"time"

	// Packages
	"github.com/djthorpe/go-pg"
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

func Test_Queue_001(t *testing.T) {
	assert := assert.New(t)
	conn := conn.Begin(t)
	defer conn.Close()

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
	t.Run("GetQueue_1", func(t *testing.T) {
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
		_, err := client.GetQueue(context.TODO(), "nonexistent_queue")
		assert.Error(err)
		assert.ErrorIs(err, httpresponse.ErrNotFound)
	})

	// Create queue then delete queue
	t.Run("DeleteQueue_1", func(t *testing.T) {
		queue, err := client.CreateQueue(context.TODO(), schema.Queue{
			Queue: "queue_name_4",
		})
		assert.NoError(err)
		assert.NotNil(queue)

		queue2, err := client.DeleteQueue(context.TODO(), queue.Queue)
		assert.NoError(err)
		assert.NotNil(queue2)
		assert.Equal(queue.Queue, queue2.Queue)
	})

	// Delete non-existent queue
	t.Run("DeleteQueue_2", func(t *testing.T) {
		_, err := client.DeleteQueue(context.TODO(), "nonexistent_queue")
		assert.Error(err)
		assert.ErrorIs(err, httpresponse.ErrNotFound)
	})

	// Update a queue name
	t.Run("UpdateQueue_1", func(t *testing.T) {
		queue, err := client.CreateQueue(context.TODO(), schema.Queue{
			Queue: "queue_name_5",
		})
		assert.NoError(err)
		assert.NotNil(queue)

		queue2, err := client.UpdateQueue(context.TODO(), queue.Queue, schema.Queue{
			Queue: "queue_name_6",
		})
		assert.NoError(err)
		assert.NotNil(queue2)
		assert.Equal("queue_name_6", queue2.Queue)
	})

	// Update a queue TTL
	t.Run("UpdateQueue_2", func(t *testing.T) {
		queue, err := client.CreateQueue(context.TODO(), schema.Queue{
			Queue: "queue_name_7",
		})
		assert.NoError(err)
		assert.NotNil(queue)

		queue2, err := client.UpdateQueue(context.TODO(), queue.Queue, schema.Queue{
			TTL: types.DurationPtr(5 * time.Hour),
		})
		assert.NoError(err)
		assert.NotNil(queue2)
		assert.NotNil(queue2.TTL)
		assert.Equal(5*time.Hour, *queue2.TTL)
	})

	// Update a queue retry
	t.Run("UpdateQueue_3", func(t *testing.T) {
		queue, err := client.CreateQueue(context.TODO(), schema.Queue{
			Queue: "queue_name_8",
		})
		assert.NoError(err)
		assert.NotNil(queue)

		queue2, err := client.UpdateQueue(context.TODO(), queue.Queue, schema.Queue{
			Retries:    types.Uint64Ptr(10),
			RetryDelay: types.DurationPtr(5 * time.Minute),
		})
		assert.NoError(err)
		assert.NotNil(queue2)
		assert.NotNil(queue2.Retries)
		assert.Equal(uint64(10), *queue2.Retries)
		assert.NotNil(queue2.RetryDelay)
		assert.Equal(5*time.Minute, *queue2.RetryDelay)
	})

	// Update non-existent queue
	t.Run("UpdateQueue_4", func(t *testing.T) {
		_, err := client.UpdateQueue(context.TODO(), "nonexistent_queue", schema.Queue{
			Queue: "nonexistent_queue",
		})
		assert.Error(err)
		assert.ErrorIs(err, httpresponse.ErrNotFound)
	})

	// List queue
	t.Run("ListQueue_1", func(t *testing.T) {
		_, err := client.CreateQueue(context.TODO(), schema.Queue{
			Queue: "queue_name_9",
		})
		assert.NoError(err)

		list, err := client.ListQueues(context.TODO(), schema.QueueListRequest{})
		assert.NoError(err)
		assert.NotNil(list)
		assert.NotNil(list.Body)
		assert.GreaterOrEqual(len(list.Body), 1)
		assert.Equal(len(list.Body), int(list.Count))
	})

	// List queue
	t.Run("ListQueue_2", func(t *testing.T) {
		_, err := client.CreateQueue(context.TODO(), schema.Queue{
			Queue: "queue_name_10",
		})
		assert.NoError(err)

		list, err := client.ListQueues(context.TODO(), schema.QueueListRequest{
			OffsetLimit: pg.OffsetLimit{
				Limit: types.Uint64Ptr(0),
			},
		})
		assert.NoError(err)
		assert.NotNil(list)
		assert.NotNil(list.Body)
		assert.GreaterOrEqual(int(list.Count), 1)
		assert.Equal(len(list.Body), 0)
	})
}

func Test_Ticker_001(t *testing.T) {
	assert := assert.New(t)
	conn := conn.Begin(t)
	defer conn.Close()

	// Create pgqueue
	client, err := pgqueue.New(context.TODO(), conn.PoolConn, pgqueue.OptWorker(t.Name()))
	assert.NoError(err)
	assert.NotNil(client)

	// Create ticker which is disabled (NULL interval)
	t.Run("CreateTicker_1", func(t *testing.T) {
		ticker, err := client.CreateTicker(context.TODO(), schema.TickerMeta{
			Ticker: "ticker_1",
		})
		assert.NoError(err)
		assert.NotNil(ticker)
		assert.NotNil(ticker.Ticker)
		assert.Equal("ticker_1", ticker.Ticker)
		assert.Nil(ticker.Interval)
	})

	// Create ticker with 1 minute interval
	t.Run("CreateTicker_2", func(t *testing.T) {
		ticker, err := client.CreateTicker(context.TODO(), schema.TickerMeta{
			Ticker:   "ticker_2",
			Interval: types.DurationPtr(1 * time.Minute),
		})
		assert.NoError(err)
		assert.NotNil(ticker)
		assert.NotNil(ticker.Ticker)
		assert.Equal("ticker_2", ticker.Ticker)
		assert.Equal(1*time.Minute, types.PtrDuration(ticker.Interval))
	})

	// Get ticker
	t.Run("GetTicker_1", func(t *testing.T) {
		ticker, err := client.CreateTicker(context.TODO(), schema.TickerMeta{
			Ticker:   "ticker_3",
			Interval: types.DurationPtr(1 * time.Minute),
		})
		assert.NoError(err)
		assert.NotNil(ticker)

		ticker2, err := client.GetTicker(context.TODO(), ticker.Ticker)
		assert.NoError(err)
		assert.NotNil(ticker2)
		assert.NotNil(ticker2.Ticker)
		assert.Equal(ticker.Ticker, ticker2.Ticker)
		assert.Equal(ticker.Interval, ticker2.Interval)
	})

	// Get non-existent ticker
	t.Run("GetTicker_2", func(t *testing.T) {
		ticker, err := client.GetTicker(context.TODO(), "nonexistent_ticker")
		assert.Error(err)
		assert.ErrorIs(err, httpresponse.ErrNotFound)
		assert.Nil(ticker)
	})

	// Update ticker - non-existent ticker
	t.Run("UpdateTicker_1", func(t *testing.T) {
		ticker, err := client.UpdateTicker(context.TODO(), "nonexistent_ticker", schema.TickerMeta{
			Ticker: "ticker_4",
		})
		assert.Error(err)
		assert.ErrorIs(err, httpresponse.ErrNotFound)
		assert.Nil(ticker)
	})

	// Update ticker - name and interval
	t.Run("UpdateTicker_2", func(t *testing.T) {
		ticker, err := client.CreateTicker(context.TODO(), schema.TickerMeta{
			Ticker:   "ticker_5",
			Interval: types.DurationPtr(1 * time.Minute),
		})
		assert.NoError(err)
		assert.NotNil(ticker)

		ticker2, err := client.UpdateTicker(context.TODO(), ticker.Ticker, schema.TickerMeta{
			Ticker:   "ticker_6",
			Interval: types.DurationPtr(2 * time.Minute),
		})
		assert.NoError(err)
		assert.Equal("ticker_6", ticker2.Ticker)
		assert.Equal(2*time.Minute, types.PtrDuration(ticker2.Interval))
	})

	// Delete ticker - non-existent ticker
	t.Run("DeleteTicker_1", func(t *testing.T) {
		ticker, err := client.DeleteTicker(context.TODO(), "nonexistent_ticker")
		assert.Error(err)
		assert.ErrorIs(err, httpresponse.ErrNotFound)
		assert.Nil(ticker)
	})

	// Delete ticker
	t.Run("DeleteTicker_2", func(t *testing.T) {
		ticker, err := client.CreateTicker(context.TODO(), schema.TickerMeta{
			Ticker:   "ticker_7",
			Interval: types.DurationPtr(1 * time.Minute),
		})
		assert.NoError(err)
		assert.NotNil(ticker)

		ticker2, err := client.DeleteTicker(context.TODO(), ticker.Ticker)
		assert.NoError(err)
		assert.NotNil(ticker2)
		assert.Equal(ticker.Ticker, ticker2.Ticker)

		_, err = client.GetTicker(context.TODO(), ticker.Ticker)
		assert.Error(err)
		assert.ErrorIs(err, httpresponse.ErrNotFound)
	})

	// List tickers
	t.Run("ListTicker_1", func(t *testing.T) {
		ticker, err := client.CreateTicker(context.TODO(), schema.TickerMeta{
			Ticker:   "ticker_8",
			Interval: types.DurationPtr(1 * time.Minute),
		})
		assert.NoError(err)
		assert.NotNil(ticker)

		tickers, err := client.ListTickers(context.TODO(), schema.TickerListRequest{})
		assert.NoError(err)
		assert.NotNil(tickers)
		assert.NotNil(tickers.Body)
		assert.GreaterOrEqual(len(tickers.Body), 1)
		assert.Equal(len(tickers.Body), int(tickers.Count))
	})

	// Next ticker
	t.Run("NextTicker_1", func(t *testing.T) {
		ticker, err := client.CreateTicker(context.TODO(), schema.TickerMeta{
			Ticker:   "ticker_9",
			Interval: types.DurationPtr(1 * time.Minute),
		})
		assert.NoError(err)
		assert.NotNil(ticker)

		ticker2, err := client.NextTicker(context.TODO())
		assert.NoError(err)
		assert.NotNil(ticker2)
	})
}
