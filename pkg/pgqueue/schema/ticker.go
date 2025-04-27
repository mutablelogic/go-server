package schema

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	// Packages
	pg "github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type TickerName string

type TickerMeta struct {
	Ticker   string         `json:"ticker" arg:"" help:"Ticker name"`
	Interval *time.Duration `json:"interval,omitempty" help:"Interval (default 1 minute)"`
}

type Ticker struct {
	TickerMeta
	Namespace string     `json:"namespace,omitempty" help:"Namespace"`
	Ts        *time.Time `json:"timestamp,omitempty"`
}

type TickerListRequest struct {
	pg.OffsetLimit
}

type TickerList struct {
	TickerListRequest
	Count uint64   `json:"count"`
	Body  []Ticker `json:"body,omitempty"`
}

type TickerNext struct{}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t Ticker) String() string {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (t TickerMeta) String() string {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (t TickerList) String() string {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// SELECTOR

func (q TickerName) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Set ticker name
	if ticker, err := q.tickerName(); err != nil {
		return "", err
	} else {
		bind.Set("id", ticker)
	}

	switch op {
	case pg.Get:
		return tickerGet, nil
	case pg.Update:
		return tickerPatch, nil
	case pg.Delete:
		return tickerDelete, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("Unsupported TickerName operation %q", op)
	}
}

func (t TickerListRequest) Select(bind *pg.Bind, op pg.Op) (string, error) {
	t.OffsetLimit.Bind(bind, TickerListLimit)

	switch op {
	case pg.List:
		return tickerList, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("Unsupported TickerListRequest operation %q", op)
	}
}

func (t TickerNext) Select(bind *pg.Bind, op pg.Op) (string, error) {
	switch op {
	case pg.Get:
		return tickerNext, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("Unsupported TickerNext operation %q", op)
	}
}

////////////////////////////////////////////////////////////////////////////////
// READER

func (r *Ticker) Scan(row pg.Row) error {
	return row.Scan(&r.Ticker, &r.Interval, &r.Namespace, &r.Ts)
}

// TickerList
func (l *TickerList) Scan(row pg.Row) error {
	var ticker Ticker
	if err := ticker.Scan(row); err != nil {
		return err
	}
	l.Body = append(l.Body, ticker)
	return nil
}

// TickerListCount
func (l *TickerList) ScanCount(row pg.Row) error {
	return row.Scan(&l.Count)
}

////////////////////////////////////////////////////////////////////////////////
// WRITER

func (w TickerMeta) Insert(bind *pg.Bind) (string, error) {
	// Set ticker
	if ticker, err := TickerName(w.Ticker).tickerName(); err != nil {
		return "", err
	} else {
		bind.Set("id", ticker)
	}

	// Return the query
	return tickerInsert, nil
}

func (w TickerMeta) Update(bind *pg.Bind) error {
	bind.Del("patch")

	// Check for id
	if !bind.Has("id") {
		return httpresponse.ErrBadRequest.With("Missing id parameter")
	}

	// Set interval
	if w.Interval != nil {
		bind.Append("patch", `"interval" = `+bind.Set("interval", w.Interval))
	}

	// Set name
	if w.Ticker != "" {
		if name, err := TickerName(w.Ticker).tickerName(); err != nil {
			return err
		} else {
			bind.Append("patch", `"ticker" = `+bind.Set("ticker", name))
		}
	}

	// Set patch
	if patch := bind.Join("patch", ","); patch == "" {
		return httpresponse.ErrBadRequest.With("No patch values")
	} else {
		bind.Set("patch", patch)
	}

	// Return success
	return nil
}

// Normalize ticker name
func (q TickerName) tickerName() (string, error) {
	if name := strings.ToLower(strings.TrimSpace(string(q))); name == "" {
		return "", httpresponse.ErrBadRequest.With("Missing ticker name")
	} else if !types.IsIdentifier(name) {
		return "", httpresponse.ErrBadRequest.With("Invalid ticker name")
	} else {
		return name, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// SQL

// Create objects in the schema
func bootstrapTicker(ctx context.Context, conn pg.Conn) error {
	q := []string{
		tickerCreateTable,
		tickerNextFunc,
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
	tickerNextFunc = `
        -- Return the next matured ticker for a namespace
        CREATE OR REPLACE FUNCTION ${"schema"}.ticker_next(namespace TEXT) RETURNS TABLE (
            "ticker" TEXT, "interval" INTERVAL, "ns" TEXT, "ts" TIMESTAMP
        ) AS $$
			WITH 
				next_ticker AS (` + tickerSelect + `WHERE "ns" = namespace AND ("ts" IS NULL OR "ts" + "interval" <= NOW()) ORDER BY "ts" NULLS FIRST)
			UPDATE
				${"schema"}.ticker
			SET
				"ts" = TIMEZONE('UTC', NOW())
			WHERE
				"ns" = namespace AND "ticker" = (SELECT "ticker" FROM next_ticker LIMIT 1 FOR UPDATE SKIP LOCKED)
			RETURNING		
				"ticker", "interval", "ns", "ts"
        $$ LANGUAGE SQL
    `
	tickerInsert = `
		INSERT INTO ${"schema"}.ticker 
			("ns", "ticker", "interval", "ts") 
		VALUES 
			(@ns, @id, DEFAULT, DEFAULT)
		RETURNING
			"ticker", "interval", "ns", "ts"
	`
	tickerPatch = `
		UPDATE ${"schema"}.ticker SET
			${patch},
			"ts" = NULL
		WHERE
			"ns" = @ns AND "ticker" = @id
		RETURNING
			"ticker", "interval", "ns", "ts"
	`
	tickerDelete = `
		DELETE FROM 
			${"schema"}.ticker
		WHERE 
			"ns" = @ns AND "ticker" = @id
		RETURNING
			"ticker", "interval",  "ns", "ts"
	`
	tickerSelect = `
		SELECT
			"ticker", "interval", "ns", "ts"
		FROM
			${"schema"}.ticker
	`
	tickerList = tickerSelect + ` WHERE "ns" = @ns`
	tickerGet  = tickerList + ` AND "ticker" = @id`
	tickerNext = `
		SELECT * FROM ${"schema"}.ticker_next(@ns)
	`
)
