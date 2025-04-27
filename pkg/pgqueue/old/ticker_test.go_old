package pgqueue_test

import (
	"context"
	"testing"
	"time"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
	assert "github.com/stretchr/testify/assert"
)

func Test_Ticker(t *testing.T) {
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
