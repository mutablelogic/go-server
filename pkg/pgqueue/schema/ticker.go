package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	// Packages
	pg "github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type TickerMeta struct {
	Ticker   string         `json:"ticker"`
	Interval *time.Duration `json:"interval,omitempty"`
}

type Ticker struct {
	TickerMeta
	Ts *time.Time `json:"timestamp,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t Ticker) String() string {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// SELECTOR

func (t TickerMeta) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Set ticker unless op is List
	if op != pg.List {
		if ticker := strings.TrimSpace(t.Ticker); ticker == "" {
			return "", httpresponse.ErrBadRequest.With("Ticker name is required")
		} else {
			bind.Set("ticker", ticker)
		}
	}

	// Return the query
	switch op {
	case pg.Get:
		return tickerGet, nil
	case pg.Delete:
		return tickerDelete, nil
	case pg.List:
		return tickerNext, nil
	case pg.Update:
		return tickerPatch, nil
	default:
		return "", fmt.Errorf("unsupported TickerMeta operation: %v", op)
	}
}

////////////////////////////////////////////////////////////////////////////////
// READER

func (r *Ticker) Scan(row pg.Row) error {
	return row.Scan(&r.Ticker, &r.Interval, &r.Ts)
}

////////////////////////////////////////////////////////////////////////////////
// WRITER

func (w TickerMeta) Insert(bind *pg.Bind) (string, error) {
	// Set ticker
	if ticker := strings.TrimSpace(w.Ticker); ticker == "" {
		return "", fmt.Errorf("Ticker name is required")
	} else if !types.IsIdentifier(ticker) {
		return "", fmt.Errorf("Ticker is not an identifier")
	} else {
		bind.Set("ticker", ticker)
	}

	// Set interval
	if w.Interval != nil {
		bind.Set("interval", *w.Interval)
	} else {
		bind.Set("interval", nil)
	}

	// Return the query
	return tickerInsert, nil
}

func (w TickerMeta) Update(bind *pg.Bind) error {
	var patch []string

	// Set interval
	if w.Interval != nil {
		patch = append(patch, `"interval" = `+bind.Set("interval", w.Interval))
	}

	// Set patch
	if len(patch) == 0 {
		return httpresponse.ErrBadRequest.With("No patch values")
	} else {
		bind.Set("patch", strings.Join(patch, ","))
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// SQL

// Create objects in the schema
func bootstrapTicker(ctx context.Context, conn pg.Conn) error {
	q := []string{
		tickerCreateTable,
	}
	for _, query := range q {
		if err := conn.Exec(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

const (
	tickerCreateTable = `
		CREATE TABLE IF NOT EXISTS ${"schema"}.ticker (
			-- ticker namespace and name
			"ns"                   TEXT NOT NULL,
			"ticker"               TEXT NOT NULL,
			-- interval (NULL means disabled)
			"interval"             INTERVAL DEFAULT INTERVAL '1 minute',
			-- last tick
			"ts"                   TIMESTAMP,
			-- primary key
			PRIMARY KEY ("ns", "ticker")
		)
	`
	tickerInsert = `
		INSERT INTO ${"schema"}.ticker 
			("ns", "ticker", "interval", "ts") 
		VALUES 
			(@ns, @ticker, @interval, DEFAULT)
		RETURNING
			"ticker", "interval", "ts"
	`
	tickerPatch = `
		UPDATE ${"schema"}.ticker SET
			${patch},
			"ts" = NULL
		WHERE
			"ns" = @ns AND "ticker" = @ticker
		RETURNING
			"ns", "ticker", "interval", "ts"
	`
	tickerDelete = `
		DELETE FROM 
			${"schema"}.ticker
		WHERE 
			"ns" = @ns AND "ticker" = @ticker
		RETURNING
			"ns", "ticker", "interval", "ts"
	`
	tickerSelect = `
		SELECT
			"ns", "ticker", "interval", "ts"
		FROM
			${"schema"}.ticker
	`
	tickerGet  = tickerSelect + `WHERE "ns" = @ns AND "ticker" = @ticker`
	tickerNext = `
		WITH 
			next_ticker AS (` + tickerSelect + `WHERE "ns" = @ns AND ("ts" IS NULL OR "ts" + "interval" < TIMEZONE('UTC', NOW())))
		UPDATE
			${"schema"}.ticker
		SET
			"ts" = TIMEZONE('UTC', NOW())
		WHERE
			"ns" = @ns AND "ticker" = (SELECT "ticker" FROM next_ticker ORDER BY "ts" LIMIT 1 FOR UPDATE SKIP LOCKED)
		RETURNING		
			"ns", "ticker", "interval", "ts"
	`
)
