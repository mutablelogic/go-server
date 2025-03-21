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

type Queue struct {
	Queue      string         `json:"queue,omitempty"`
	TTL        *time.Duration `json:"ttl,omitempty"`
	Retries    *uint64        `json:"retries"`
	RetryDelay *time.Duration `json:"retry_delay"`
}

type QueueListRequest struct {
	pg.OffsetLimit
}

type QueueList struct {
	QueueListRequest
	Count uint64  `json:"count"`
	Body  []Queue `json:"body,omitempty"`
}

type QueueCleanRequest struct{}

type QueueCleanResponse struct {
	Body []Task `json:"body,omitempty"`
}

type QueueStatus struct {
	Queue  string `json:"queue"`
	Status string `json:"status"`
	Count  uint64 `json:"count"`
}

type QueueStatusRequest struct{}

type QueueStatusResponse struct {
	Body []QueueStatus `json:"body,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (q Queue) String() string {
	data, err := json.MarshalIndent(q, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (q QueueList) String() string {
	data, err := json.MarshalIndent(q, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (q QueueStatus) String() string {
	data, err := json.MarshalIndent(q, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// READER

// Queue
func (q *Queue) Scan(row pg.Row) error {
	return row.Scan(&q.Queue, &q.TTL, &q.Retries, &q.RetryDelay)
}

// QueueList
func (l *QueueList) Scan(row pg.Row) error {
	var queue Queue
	if err := queue.Scan(row); err != nil {
		return err
	}
	l.Body = append(l.Body, queue)
	return nil
}

// QueueListCount
func (l *QueueList) ScanCount(row pg.Row) error {
	return row.Scan(&l.Count)
}

// QueueCleanResponse
func (l *QueueCleanResponse) Scan(row pg.Row) error {
	var task Task
	if err := task.Scan(row); err != nil {
		return err
	}
	l.Body = append(l.Body, task)
	return nil
}

// QueueStatus
func (s *QueueStatus) Scan(row pg.Row) error {
	return row.Scan(&s.Queue, &s.Status, &s.Count)
}

// QueueStatusResponse
func (l *QueueStatusResponse) Scan(row pg.Row) error {
	var status QueueStatus
	if err := status.Scan(row); err != nil {
		return err
	}
	l.Body = append(l.Body, status)
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// SELECTOR

func (q Queue) Select(bind *pg.Bind, op pg.Op) (string, error) {
	switch op {
	case pg.Get:
		return queueGet, nil
	case pg.Update:
		return queuePatch, nil
	case pg.Delete:
		return queueDelete, nil
	default:
		return "", fmt.Errorf("Unsupported Queue operation %q", op)
	}
}

func (q QueueCleanRequest) Select(bind *pg.Bind, op pg.Op) (string, error) {
	switch op {
	case pg.List:
		return queueClean, nil
	default:
		return "", fmt.Errorf("Unsupported QueueCleanRequest operation %q", op)
	}
}

func (l QueueListRequest) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Bind parameters
	bind.Set("where", "")
	l.OffsetLimit.Bind(bind, QueueListLimit)

	switch op {
	case pg.List:
		return queueList, nil
	default:
		return "", fmt.Errorf("Unsupported QueueListRequest operation %q", op)
	}
}

func (l QueueStatusRequest) Select(bind *pg.Bind, op pg.Op) (string, error) {
	var where []string

	// Bind parameters
	if bind.Has("id") {
		where = append(where, `queue = @id`)
	}

	// Where clause
	if len(where) > 0 {
		bind.Set("where", "WHERE "+strings.Join(where, " AND "))
	} else {
		bind.Set("where", "")
	}

	switch op {
	case pg.List:
		return queueStats, nil
	default:
		return "", fmt.Errorf("Unsupported QueueStatusRequest operation %q", op)
	}
}

////////////////////////////////////////////////////////////////////////////////
// WRITER

// Insert
func (q Queue) Insert(bind *pg.Bind) (string, error) {
	// Queue name
	queue, err := q.queueName()
	if err != nil {
		return "", err
	} else {
		bind.Set("queue", queue)
	}

	// Return the insert query
	return queueInsert, nil
}

// Patch
func (q Queue) Patch(bind *pg.Bind) error {
	var patch []string

	// Queue name
	if queue, err := q.queueName(); err != nil {
		return err
	} else {
		patch = append(patch, `queue=`+bind.Set("queue", queue))
	}

	// Set patch values
	if q.TTL != nil {
		patch = append(patch, `ttl=`+bind.Set("ttl", q.TTL))
	}
	if q.Retries != nil {
		patch = append(patch, `retries=`+bind.Set("retries", q.Retries))
	}
	if q.RetryDelay != nil {
		patch = append(patch, `retry_delay=`+bind.Set("retry_delay", q.RetryDelay))
	}

	// Check patch values
	if len(patch) == 0 {
		return httpresponse.ErrBadRequest.With("No patch values")
	} else {
		bind.Set("patch", strings.Join(patch, ", "))
	}

	// Return success
	return nil
}

// Normalize queue name
func (q Queue) queueName() (string, error) {
	if queue := strings.ToLower(strings.TrimSpace(q.Queue)); queue == "" {
		return "", httpresponse.ErrBadRequest.With("Missing queue name")
	} else if !types.IsIdentifier(queue) {
		return "", httpresponse.ErrBadRequest.With("Invalid queue name")
	} else {
		return queue, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// SQL

// Create objects in the schema
func bootstrapQueue(ctx context.Context, conn pg.Conn) error {
	q := []string{
		queueCreateTable,
	}
	for _, query := range q {
		if err := conn.Exec(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

const (
	queueCreateTable = `
		CREATE TABLE IF NOT EXISTS ${"schema"}.queue (
			-- queue name
			"queue" TEXT PRIMARY KEY,
			-- time-to-live for queue messages
			"ttl" INTERVAL DEFAULT INTERVAL '1 hour',
			-- number of retries before failing
			"retries" INTEGER NOT NULL DEFAULT 3 CHECK ("retries" >= 0),
			-- delay between retries in seconds
			"retry_delay" INTERVAL NOT NULL DEFAULT INTERVAL '2 minute'
		)
	`
	queueInsert = `
		INSERT INTO ${"schema"}.queue (
			queue, ttl, retries, retry_delay
		) VALUES (
		 	@queue, DEFAULT, DEFAULT, DEFAULT
		) RETURNING 
			queue, ttl, retries, retry_delay
	`
	queueGet = `
		SELECT
			queue, ttl, retries, retry_delay
		FROM
			${"schema"}.queue
		WHERE
			queue = @id
	`
	queuePatch = `
		UPDATE ${"schema"}.queue SET
			${patch}
		WHERE
			queue = @id
		RETURNING
			queue, ttl, retries, retry_delay
	`
	queueDelete = `
		DELETE FROM ${"schema"}.queue WHERE
			queue = @id
		RETURNING
			queue, ttl, retries, retry_delay
	`
	queueList = `
		SELECT
			queue, ttl, retries, retry_delay
		FROM 
			${"schema"}.queue ${where}
	`
	queueClean = `
		SELECT * FROM ${"schema"}.queue_clean(@id)
	`
	queueStatsView = `
		CREATE OR REPLACE VIEW ${"schema"}."queue_status" AS
		SELECT 
			Q."queue", S."status", T."count"
		FROM
			${"schema"}."queue" Q
		CROSS JOIN
			(SELECT UNNEST(enum_range(NULL::${"schema"}.STATUS)) AS "status") S
		LEFT JOIN
			(SELECT "id", "queue", ${"schema"}.queue_task_status(id) AS "status" FROM ${"schema"}."task") T
		ON
			S."status" = T."status" AND Q."queue" = T."queue"
		GROUP BY
			1, 2
	`
	queueStats = `
		SELECT "queue", "status", "count" FROM ${"schema"}.queue_status ${where}
	`
)
