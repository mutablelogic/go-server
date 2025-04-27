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

type TaskId uint64

type TaskRetain struct {
	Worker string `json:"worker,omitempty"`
}

type TaskRelease struct {
	Id     uint64 `json:"id,omitempty"`
	Fail   bool   `json:"fail,omitempty"`
	Result any    `json:"result,omitempty"`
}

type TaskMeta struct {
	Payload   any        `json:"payload,omitempty"`
	DelayedAt *time.Time `json:"delayed_at,omitempty"`
}

type Task struct {
	Id uint64 `json:"id,omitempty"`
	TaskMeta
	Worker     *string    `json:"worker,omitempty"`
	Namespace  string     `json:"namespace,omitempty"`
	Queue      string     `json:"queue,omitempty"`
	Result     any        `json:"result,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	DiesAt     *time.Time `json:"dies_at,omitempty"`
	Retries    *uint64    `json:"retries,omitempty"`
}

type TaskWithStatus struct {
	Task
	Status string `json:"status,omitempty"`
}

type TaskListRequest struct {
	pg.OffsetLimit
	Status string `json:"status,omitempty"`
}

type TaskList struct {
	TaskListRequest
	Count uint64           `json:"count"`
	Body  []TaskWithStatus `json:"body,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t Task) String() string {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (t TaskMeta) String() string {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (t TaskWithStatus) String() string {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (t TaskList) String() string {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// READER

func (t *TaskId) Scan(row pg.Row) error {
	var id *uint64
	if err := row.Scan(&id); err != nil {
		return err
	} else {
		*t = TaskId(types.PtrUint64(id))
	}
	return nil
}

func (t *Task) Scan(row pg.Row) error {
	return row.Scan(&t.Id, &t.Queue, &t.Namespace, &t.Payload, &t.Result, &t.Worker, &t.CreatedAt, &t.DelayedAt, &t.StartedAt, &t.FinishedAt, &t.DiesAt, &t.Retries)
}

func (t *TaskWithStatus) Scan(row pg.Row) error {
	return row.Scan(&t.Id, &t.Queue, &t.Namespace, &t.Payload, &t.Result, &t.Worker, &t.CreatedAt, &t.DelayedAt, &t.StartedAt, &t.FinishedAt, &t.DiesAt, &t.Retries, &t.Status)
}

// TaskList
func (l *TaskList) Scan(row pg.Row) error {
	var task TaskWithStatus
	if err := task.Scan(row); err != nil {
		return err
	}
	l.Body = append(l.Body, task)
	return nil
}

// TaskListCount
func (l *TaskList) ScanCount(row pg.Row) error {
	return row.Scan(&l.Count)
}

////////////////////////////////////////////////////////////////////////////////
// WRITER

func (t TaskMeta) Insert(bind *pg.Bind) (string, error) {
	if !bind.Has("id") {
		return "", fmt.Errorf("missing queue id")
	} else {
		bind.Set("queue", bind.Get("id"))
	}
	if t.Payload == nil {
		return "", fmt.Errorf("missing payload")
	} else if data, err := json.Marshal(t.Payload); err != nil {
		return "", err
	} else {
		bind.Set("payload", string(data))
	}
	if t.DelayedAt != nil {
		if t.DelayedAt.Before(time.Now()) {
			return "", fmt.Errorf("delayed_at is in the past")
		}
		bind.Set("delayed_at", t.DelayedAt.UTC())
	} else {
		bind.Set("delayed_at", nil)
	}
	return taskInsert, nil
}

func (t TaskMeta) Update(bind *pg.Bind) error {
	bind.Del("patch")

	// DelayedAt
	if t.DelayedAt != nil {
		if t.DelayedAt.Before(time.Now()) {
			return fmt.Errorf("delayed_at is in the past")
		}
		bind.Append("patch", `delayed_at = `+bind.Set("delayed_at", t.DelayedAt))
	}

	// Payload
	if t.Payload != nil {
		data, err := json.Marshal(t.Payload)
		if err != nil {
			return err
		}
		bind.Append("patch", `payload = `+bind.Set("payload", string(data)))
	}

	// Set patch
	if patch := bind.Join("patch", ", "); patch == "" {
		return fmt.Errorf("no fields to update")
	} else {
		bind.Set("patch", patch)
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// SELECTOR

func (t TaskId) Select(bind *pg.Bind, op pg.Op) (string, error) {
	bind.Set("tid", t)
	switch op {
	case pg.Get:
		return taskGet, nil
	default:
		return "", fmt.Errorf("unsupported TaskId operation %q", op)
	}
}

func (l TaskListRequest) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Bind parameters
	var where []string
	if l.Status != "" {
		where = append(where, `status=`+bind.Set("status", l.Status))
	}
	if len(where) == 0 {
		bind.Set("where", "")
	} else {
		bind.Set("where", "WHERE "+strings.Join(where, " AND "))
	}
	l.OffsetLimit.Bind(bind, QueueListLimit)

	switch op {
	case pg.List:
		return taskList, nil
	default:
		return "", fmt.Errorf("unsupported TaskListRequest operation %q", op)
	}
}

func (t TaskRetain) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Worker is required
	if worker := strings.TrimSpace(t.Worker); worker == "" {
		return "", httpresponse.ErrBadRequest.Withf("Missing worker")
	} else {
		bind.Set("worker", worker)
	}

	// Retain
	switch op {
	case pg.Get:
		return taskRetain, nil
	default:
		return "", fmt.Errorf("unsupported TaskRetain operation %q", op)
	}
}

func (t TaskRelease) Select(bind *pg.Bind, op pg.Op) (string, error) {
	if t.Id == 0 {
		return "", httpresponse.ErrBadRequest.Withf("Missing task id")
	} else {
		bind.Set("tid", t.Id)
	}

	// Result of the task
	data, err := json.Marshal(t.Result)
	if err != nil {
		return "", err
	} else {
		bind.Set("result", string(data))
	}

	// Release
	switch op {
	case pg.Get:
		if t.Fail {
			return taskFail, nil
		} else {
			return taskRelease, nil
		}
	default:
		return "", fmt.Errorf("unsupported TaskRelease operation %q", op)
	}
}

////////////////////////////////////////////////////////////////////////////////
// SQL

// Create objects in the schema
func bootstrapTask(ctx context.Context, conn pg.Conn) error {
	q := []string{
		taskCreateTable,
		taskCreateInsertFunc,
		taskCreateNotifyFunc,
		taskCreateTrigger,
		taskStatusType,
		taskStatusFunc,
		taskBackoffFunc,
		taskRetainFunc,
		taskReleaseFunc,
		taskFailFunc,
		taskCleanFunc,
		queueStatsView,
	}
	for _, query := range q {
		if err := conn.Exec(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

const (
	taskCreateTable = `
        CREATE TABLE IF NOT EXISTS ${"schema"}."task" (
            -- task id
            "id"                         SERIAL PRIMARY KEY,
            -- namespace and queue
			"ns"                         TEXT NOT NULL,
            "queue"                      TEXT NOT NULL,
            -- task payload and result
            -- result is used for both successful and unsuccessful tasks
            "payload"                    JSONB NOT NULL DEFAULT '{}',
            "result"                     JSONB NOT NULL DEFAULT 'null',
            -- worker identifier
            "worker"                     TEXT,
            -- when the task has been created
            "created_at"                 TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
            -- when the task should be next run (or NULL if it should be run now)
            "delayed_at"                 TIMESTAMP,
            -- when the task has been started, finished and when it expires (dies)
            "started_at"                 TIMESTAMP,
            "finished_at"                TIMESTAMP,
            "dies_at"                    TIMESTAMP,
            -- task maximum retries. when this reaches zero the task is marked as failed
            "retries"                    INTEGER NOT NULL,
            -- the initial retry value
            "initial_retries"            INTEGER NOT NULL,
			-- foreign key (ns, queue) references queue (ns, queue)
			FOREIGN KEY ("ns", "queue") REFERENCES ${"schema"}.queue ("ns", "queue") ON DELETE CASCADE
        )
    `
	taskCreateInsertFunc = `
        -- Insert a new payload into a queue
        CREATE OR REPLACE FUNCTION ${"schema"}.queue_insert(n TEXT, q TEXT, p JSONB, delayed_at TIMESTAMP) RETURNS BIGINT AS $$
        WITH defaults AS (
            -- Select the retries and ttl from the queue defaults
            SELECT
                retries, CASE 
                    WHEN "ttl" IS NULL THEN NULL 
                    WHEN "delayed_at" IS NULL OR "delayed_at" < TIMEZONE('UTC', NOW()) THEN TIMEZONE('UTC', NOW()) + "ttl"                    
                    ELSE "delayed_at" + "ttl" 
                END AS dies_at
            FROM
                ${"schema"}."queue"
            WHERE
                "queue" = q AND "ns" = n
            LIMIT
                1
        ) INSERT INTO 
            ${"schema"}."task" ("ns", "queue", "payload", "delayed_at", "retries", "initial_retries", "dies_at")
        SELECT
            n, q, p, CASE
                WHEN "delayed_at" IS NULL THEN NULL
                WHEN "delayed_at" < TIMEZONE('UTC', NOW()) THEN (NOW() AT TIME ZONE 'UTC')
                ELSE "delayed_at"
            END, retries, retries, dies_at
        FROM
            defaults
        RETURNING
            "id"
        $$ LANGUAGE SQL
    `
	taskCreateNotifyFunc = `
        CREATE OR REPLACE FUNCTION ${"schema"}.queue_notify() RETURNS TRIGGER AS $$
        BEGIN
            PERFORM pg_notify(NEW.ns || '_queue_insert',NEW.queue);
            RETURN NEW;
        END;
        $$ LANGUAGE plpgsql
    `
	taskCreateTrigger = `
        CREATE OR REPLACE TRIGGER 
            queue_insert_tigger 
        AFTER INSERT ON ${"schema"}."task" FOR EACH ROW EXECUTE FUNCTION 
            ${"schema"}.queue_notify()
    `
	taskInsert = `
        -- Insert a new task into a queue and return the id
        SELECT ${"schema"}.queue_insert(@ns, @queue, @payload, @delayed_at)
    `
	taskRetainFunc = `
        -- A specific worker locks a task in a queue for processing
        CREATE OR REPLACE FUNCTION ${"schema"}.queue_lock(n TEXT, w TEXT) RETURNS BIGINT AS $$
        UPDATE ${"schema"}."task" SET 
            "started_at" = TIMEZONE('UTC', NOW()), "worker" = w, "result" = 'null'
        WHERE "id" = (
            SELECT 
				"id" 
            FROM
				${"schema"}."task"
            WHERE
				"ns" = n
            AND
				("started_at" IS NULL AND "finished_at" IS NULL AND "dies_at" > TIMEZONE('UTC', NOW()))
            AND 
                ("delayed_at" IS NULL OR "delayed_at" <= TIMEZONE('UTC', NOW()))
            AND
                ("retries" > 0)
            ORDER BY
                "created_at"
            FOR UPDATE SKIP LOCKED LIMIT 1
        ) RETURNING
			"id"
        $$ LANGUAGE SQL
    `
	taskReleaseFunc = `
        -- Unlock a task in a queue with successful result
        CREATE OR REPLACE FUNCTION ${"schema"}.queue_unlock(tid BIGINT, r JSONB) RETURNS BIGINT AS $$
            UPDATE ${"schema"}."task" SET 
				"finished_at" = TIMEZONE('UTC', NOW()), "dies_at" = NULL, "result" = r
            WHERE 
				("id" = tid)
            AND
				("started_at" IS NOT NULL AND "finished_at" IS NULL AND "dies_at" > TIMEZONE('UTC', NOW()))
            RETURNING
				"id"
        $$ LANGUAGE SQL
    `
	taskStatusType = `
        -- Create the status type
        DO $$ BEGIN
			CREATE TYPE ${"schema"}.STATUS AS ENUM('expired', 'new', 'failed', 'retry', 'retained', 'released', 'unknown');
        EXCEPTION
			WHEN duplicate_object THEN null;
        END $$;
    `
	taskStatusFunc = `
        -- Return the current status of a task
        CREATE OR REPLACE FUNCTION ${"schema"}.queue_task_status(t BIGINT) RETURNS ${"schema"}.STATUS AS $$
            SELECT CASE
                WHEN "dies_at" IS NOT NULL AND "dies_at" < TIMEZONE('UTC', NOW()) THEN 'expired'::${"schema"}.STATUS
                WHEN "started_at" IS NULL AND "finished_at" IS NULL  AND "retries" = "initial_retries" THEN 'new'::${"schema"}.STATUS
                WHEN "started_at" IS NULL AND "finished_at" IS NULL  AND "retries" = 0 THEN 'failed'::${"schema"}.STATUS
                WHEN "started_at" IS NULL AND "finished_at" IS NULL  THEN 'retry'::${"schema"}.STATUS
                WHEN "started_at" IS NOT NULL AND "finished_at" IS NULL  THEN 'retained'::${"schema"}.STATUS
                WHEN "started_at" IS NOT NULL AND "finished_at" IS NOT NULL THEN 'released'::${"schema"}.STATUS
                ELSE 'unknown'::${"schema"}.STATUS
            END AS "status"
            FROM 
                ${"schema"}."task"
            WHERE
              "id" = t              
        $$ LANGUAGE SQL
    `
	taskBackoffFunc = `
        -- Backoff exponentially with retries    
        CREATE OR REPLACE FUNCTION ${"schema"}.queue_backoff(tid BIGINT) RETURNS TIMESTAMP AS $$
            SELECT CASE
                WHEN T."retries" = 0 THEN NULL
                ELSE (NOW() AT TIME ZONE 'UTC') + Q."retry_delay" * (POW(T."initial_retries" - T."retries" + 1,2))
            END AS "delayed_at"
            FROM
                ${"schema"}."task" T
            JOIN            	
                ${"schema"}."queue" Q
            ON
                T."queue" = Q."queue" AND T."ns" = Q."ns"
            WHERE
                T."id" = tid
        $$ LANGUAGE SQL
    `
	taskFailFunc = `
        -- Unlock a task in a queue with fail result
        CREATE OR REPLACE FUNCTION ${"schema"}.queue_fail(tid BIGINT, r JSONB) RETURNS BIGINT AS $$
            UPDATE ${"schema"}."task" SET 
                "retries" = "retries" - 1, "result" = r, "started_at" = NULL, "finished_at" = NULL, "delayed_at" = ${"schema"}.queue_backoff(tid)
            WHERE 
                "id" = tid AND "retries" > 0 AND ("started_at" IS NOT NULL AND "finished_at" IS NULL)
            RETURNING
				"id"
        $$ LANGUAGE SQL 
    `
	taskCleanFunc = `
		-- Cleanup tasks in a queue which are in an end state 
		CREATE OR REPLACE FUNCTION ${"schema"}.queue_clean(n TEXT, q TEXT) RETURNS TABLE (
            "id" BIGINT, "queue" TEXT, "ns" TEXT, "payload" JSONB, "result" JSONB, "worker" TEXT, "created_at" TIMESTAMP, "delayed_at" TIMESTAMP, "started_at" TIMESTAMP, "finished_at" TIMESTAMP, "dies_at" TIMESTAMP, "retries" INTEGER
        ) AS $$
			DELETE FROM
				${"schema"}."task"
			WHERE
				id IN (
                    WITH sq AS (
                        SELECT 
                            "id", ${"schema"}.queue_task_status("id") AS "status" 
                        FROM 
                            ${"schema"}."task" 
                        WHERE
                            "ns" = n AND "queue" = q
                        AND
                            (dies_at IS NULL OR dies_at < TIMEZONE('UTC', NOW()))
                    ) SELECT 
					 	"id" 
					FROM 
						sq 
					WHERE 
						"status" IN ('expired', 'released', 'failed')
					ORDER BY 
						"created_at"
					LIMIT 
						100
                )
			RETURNING
				"id", "queue", "ns", "payload", "result", "worker", "created_at", "delayed_at", "started_at", "finished_at", "dies_at", "retries"
		$$ LANGUAGE SQL
	`
	taskRetain = `
        -- Returns the id of the task which has been retained
        SELECT ${"schema"}.queue_lock(@ns, @worker)
    `
	taskRelease = `
        -- Returns the id of the task which has been released
        SELECT ${"schema"}.queue_unlock(@tid, @result)
    `
	taskFail = `
        -- Returns the id of the task which has been failed
        SELECT ${"schema"}.queue_fail(@tid, @result)
    `
	taskSelect = `
        SELECT 
            "id", "queue", "ns", "payload", "result", "worker", "created_at", "delayed_at", "started_at", "finished_at", "dies_at", "retries", ${"schema"}.queue_task_status("id") AS "status"
        FROM
            ${"schema"}."task"
    `
	taskGet  = taskSelect + `WHERE "id" = @tid`
	taskList = `WITH q AS (` + taskSelect + `) SELECT * FROM q ${where}`
)
