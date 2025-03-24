package pgqueue_test

import (
	"context"
	"testing"
	"time"

	// Packages
	test "github.com/djthorpe/go-pg/pkg/test"
	server "github.com/mutablelogic/go-server"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
	pg "github.com/mutablelogic/go-server/plugin/pg"
	pgqueue "github.com/mutablelogic/go-server/plugin/pgqueue"
	assert "github.com/stretchr/testify/assert"
)

// Global connection variable
var conn test.Conn

// Start up a container and test the pool
func TestMain(m *testing.M) {
	test.Main(m, &conn)
}

func Test_task_001(t *testing.T) {
	assert := assert.New(t)
	client, err := pgqueue.Config{
		Pool: pg.NewTask(conn.PoolConn),
	}.New(context.TODO())
	if assert.NoError(err) {
		assert.NotNil(client)
	}

	t.Run("RegisterTicker_1", func(t *testing.T) {
		// Create a ticker
		err := client.(server.PGQueue).RegisterTicker(context.TODO(), schema.TickerMeta{
			Ticker:   "ticker_name_1",
			Interval: types.DurationPtr(time.Second),
		}, func(ctx context.Context, _ any) error {
			// Callback function for ticker
			t.Log("Ticker fired", pgqueue.Ticker(ctx))
			return nil
		})
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Run for 10 seconds
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Run tickers
		err = client.Run(ctx)
		if !assert.NoError(err) {
			t.SkipNow()
		}
	})

	t.Run("RegisterTicker_2", func(t *testing.T) {
		// Create a ticker
		err := client.(server.PGQueue).RegisterTicker(context.TODO(), schema.TickerMeta{
			Ticker:   "ticker_name_2",
			Interval: types.DurationPtr(time.Second),
		}, func(ctx context.Context, _ any) error {
			panic("test panic " + pgqueue.Ticker(ctx).Ticker)
		})
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Run for 10 seconds
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Run tickers
		err = client.Run(ctx)
		if !assert.NoError(err) {
			t.SkipNow()
		}
	})

	t.Run("RegisterTicker_3", func(t *testing.T) {
		// Create a ticker
		err := client.(server.PGQueue).RegisterTicker(context.TODO(), schema.TickerMeta{
			Ticker:   "ticker_name_3",
			Interval: types.DurationPtr(time.Second),
		}, func(ctx context.Context, _ any) error {
			// Callback function for ticker
			t.Log("Ticker fired", pgqueue.Ticker(ctx))
			// Sleep beyond the deadline
			time.Sleep(2 * time.Second)
			return nil
		})
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Run for 10 seconds
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Run tickers
		err = client.Run(ctx)
		if !assert.NoError(err) {
			t.SkipNow()
		}
	})
}
