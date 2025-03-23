package pgqueue_test

import (
	"context"
	"testing"
	"time"

	// Packages
	test "github.com/djthorpe/go-pg/pkg/test"
	server "github.com/mutablelogic/go-server"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	"github.com/mutablelogic/go-server/pkg/types"
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
	pgqueue, err := pgqueue.Config{
		Pool: pg.NewTask(conn.PoolConn),
	}.New(context.TODO())
	if assert.NoError(err) {
		assert.NotNil(pgqueue)
	}

	t.Run("RegisterTicker", func(t *testing.T) {
		// Create a ticker
		ticker, err := pgqueue.(server.PGQueue).RegisterTicker(context.TODO(), schema.TickerMeta{
			Ticker:   "ticker_name_1",
			Interval: types.DurationPtr(time.Second),
		})
		if !assert.NoError(err) {
			t.SkipNow()
		}
		assert.NotNil(ticker)

		// Run for 30 seconds
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Run tickers
		err = pgqueue.Run(ctx)
		if !assert.NoError(err) {
			t.SkipNow()
		}
	})
}
