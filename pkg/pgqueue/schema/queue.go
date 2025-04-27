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

type QueueName string

type QueueMeta struct {
	Queue      string         `json:"queue,omitempty" arg:"" help:"Queue name"`
	TTL        *time.Duration `json:"ttl,omitempty" help:"Time-to-live for queue messages"`
	Retries    *uint64        `json:"retries" help:"Number of retries before failing"`
	RetryDelay *time.Duration `json:"retry_delay" help:"Backoff delay"`
}

type Queue struct {
	QueueMeta
	Namespace string `json:"namespace,omitempty" help:"Namespace"`
}

type QueueListRequest struct {
	pg.OffsetLimit
}

type QueueList struct {
	QueueListRequest
	Count uint64  `json:"count"`
	Body  []Queue `json:"body,omitempty"`
}

type QueueCleanRequest struct {
	Queue string `json:"queue,omitempty" arg:"" help:"Queue name"`
}

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

func (q QueueMeta) String() string {
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
	return row.Scan(&q.Queue, &q.TTL, &q.Retries, &q.RetryDelay, &q.Namespace)
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

func (q QueueName) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Set queue name
	if name, err := q.queueName(); err != nil {
		return "", err
	} else {
		bind.Set("id", name)
	}

	switch op {
	case pg.Get:
		return queueGet, nil
	case pg.Update:
		return queuePatch, nil
	case pg.Delete:
		return queueDelete, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("Unsupported QueueName operation %q", op)
	}
}

func (q QueueCleanRequest) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Set queue name
	if name, err := QueueName(q.Queue).queueName(); err != nil {
		return "", err
	} else {
		bind.Set("id", name)
	}

	switch op {
	case pg.List:
		return queueClean, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("Unsupported QueueCleanRequest operation %q", op)
	}
}

func (l QueueListRequest) Select(bind *pg.Bind, op pg.Op) (string, error) {
	bind.Set("where", "")
	l.OffsetLimit.Bind(bind, QueueListLimit)

	switch op {
	case pg.List:
		return queueList, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("Unsupported QueueListRequest operation %q", op)
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
		return "", httpresponse.ErrInternalError.Withf("Unsupported QueueStatusRequest operation %q", op)
	}
}

////////////////////////////////////////////////////////////////////////////////
// WRITER

// Insert
func (q QueueMeta) Insert(bind *pg.Bind) (string, error) {
	// Queue name
	queue, err := QueueName(q.Queue).queueName()
	if err != nil {
		return "", err
	} else {
		bind.Set("queue", queue)
	}

	// Note: Inserts default values for ttl, retries, retry_delay
	// A subsequent update is required to set these values

	// Return the insert query
	return queueInsert, nil
}

// Patch
func (q QueueMeta) Update(bind *pg.Bind) error {
	var patch []string

	// Queue name
	if q.Queue != "" {
		if queue, err := QueueName(q.Queue).queueName(); err != nil {
			return err
		} else {
			patch = append(patch, `queue=`+bind.Set("queue", queue))
		}
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
func (q QueueName) queueName() (string, error) {
	if queue := strings.ToLower(strings.TrimSpace(string(q))); queue == "" {
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
			-- namespace and queue name
			"ns" TEXT NOT NULL,
			"queue" TEXT NOT NULL,
			-- time-to-live for queue messages
			"ttl" INTERVAL DEFAULT INTERVAL '1 hour',
			-- number of retries before failing
			"retries" INTEGER NOT NULL DEFAULT 3 CHECK ("retries" >= 0),
			-- delay between retries in seconds
			"retry_delay" INTERVAL NOT NULL DEFAULT INTERVAL '2 minute',
			-- primary key
			PRIMARY KEY ("ns", "queue")
		)
	`
	queueInsert = `
		INSERT INTO ${"schema"}.queue (
			ns, queue, ttl, retries, retry_delay
		) VALUES (
		 	@ns, @queue, DEFAULT, DEFAULT, DEFAULT
		) RETURNING 
			queue, ttl, retries, retry_delay, ns
	`
	queueGet = `
		SELECT
			queue, ttl, retries, retry_delay, ns
		FROM
			${"schema"}.queue
		WHERE
			queue = @id
		AND
			ns = @ns
	`
	queuePatch = `
		UPDATE ${"schema"}.queue SET
			${patch}
		WHERE
			queue = @id
		AND
			ns = @ns
		RETURNING
			queue, ttl, retries, retry_delay, ns
	`
	queueDelete = `
		DELETE FROM ${"schema"}.queue WHERE
			queue = @id
		AND
			ns = @ns
		RETURNING
			queue, ttl, retries, retry_delay, ns
	`
	queueList = `
		SELECT
			queue, ttl, retries, retry_delay, ns
		FROM 
			${"schema"}.queue 
		WHERE
			ns = @ns ${where}
	`
	queueClean = `
		SELECT * FROM ${"schema"}.queue_clean(@ns, @id)
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
		SELECT "queue", "status", "count" FROM ${"schema"}.queue_status WHERE "ns" = @ns ${where}
	`
)
