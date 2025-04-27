package schema

import (
	"encoding/json"
	"strings"
	"time"

	// Packages
	pg "github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ConnectionPid uint64

type Connection struct {
	Pid         uint32    `json:"pid" help:"Process ID"`
	Database    string    `json:"database" help:"Database"`
	Role        string    `json:"role" help:"Role"`
	Application *string   `json:"application,omitempty" help:"Application"`
	ClientAddr  string    `json:"client_addr,omitempty" help:"Client address"`
	ClientPort  uint16    `json:"client_port,omitempty" help:"Client port"`
	ConnStart   time.Time `json:"conn_start,omitempty" help:"Connection start"`
	QueryStart  time.Time `json:"query_start,omitempty" help:"Query start"`
	Query       string    `json:"query,omitempty" help:"Query"`
	State       string    `json:"state,omitempty" help:"State"`
}

type ConnectionListRequest struct {
	pg.OffsetLimit
	Database *string `json:"database,omitempty" help:"Database"`
	Role     *string `json:"role,omitempty" help:"Role"`
	State    *string `json:"state,omitempty" help:"State"`
}

type ConnectionList struct {
	Count uint64       `json:"count"`
	Body  []Connection `json:"body,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c Connection) String() string {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (c ConnectionListRequest) String() string {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (c ConnectionList) String() string {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// SELECT

func (c ConnectionListRequest) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Where
	bind.Del("where")
	if c.Database != nil {
		bind.Append("where", `"database" = `+bind.Set("database", strings.TrimSpace(*c.Database)))
	}
	if c.Role != nil {
		bind.Append("where", `"role" = `+bind.Set("role", strings.TrimSpace(*c.Role)))
	}
	if c.State != nil {
		bind.Append("where", `"state" = `+bind.Set("state", strings.TrimSpace(*c.State)))
	}
	if where := bind.Join("where", " AND "); where != "" {
		bind.Set("where", `WHERE `+where)
	} else {
		bind.Set("where", "")
	}

	// Offset and limit
	c.OffsetLimit.Bind(bind, ConnectionListLimit)

	// Return query
	switch op {
	case pg.List:
		return connectionList, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("unsupported ConnectionListRequest operation %q", op)
	}
}

func (c ConnectionPid) Select(bind *pg.Bind, op pg.Op) (string, error) {
	if c == 0 {
		return "", httpresponse.ErrBadRequest.With("missing pid")
	} else {
		bind.Set("pid", c)
	}

	// Return query
	switch op {
	case pg.Get:
		return connectionGet, nil
	case pg.Delete:
		return connectionDelete, nil
	default:
		return "", httpresponse.ErrInternalError.Withf("unsupported ConnectionListRequest operation %q", op)
	}
}

////////////////////////////////////////////////////////////////////////////////
// READER

func (c *Connection) Scan(row pg.Row) error {
	var result bool
	return row.Scan(&c.Pid, &c.Database, &c.Role, &c.Application, &c.ClientAddr, &c.ClientPort, &c.ConnStart, &c.QueryStart, &c.Query, &c.State, &result)
}

func (c *ConnectionList) Scan(row pg.Row) error {
	var connection Connection
	if err := connection.Scan(row); err != nil {
		return err
	} else {
		c.Body = append(c.Body, connection)
	}
	return nil
}

func (c *ConnectionList) ScanCount(row pg.Row) error {
	return row.Scan(&c.Count)
}

////////////////////////////////////////////////////////////////////////////////
// SQL

const (
	connectionSelect = `
		WITH conn AS (
			SELECT
				C.pid AS "pid", 
				C.datname AS "database",
				C.usename AS "role",
				NULLIF(C.application_name, '') AS "application",
				COALESCE(C.client_hostname, abbrev(C.client_addr)) AS "client_addr",
				C.client_port AS "client_port",
				C.backend_start AS "conn_start",
				C.query_start AS "query_start",
				C.query AS "query",
				C.state AS "state"
			FROM
				${"schema"}."pg_stat_activity" C
			WHERE
				C.datname IS NOT NULL
			AND
				C.state IS NOT NULL
		) SELECT * FROM conn`
	connectionGet    = `WITH q AS (` + connectionSelect + `) SELECT *, false FROM q WHERE "pid" = @pid`
	connectionList   = `WITH q AS (` + connectionSelect + `) SELECT *, false FROM q ${where}`
	connectionDelete = `WITH q AS (` + connectionSelect + `) SELECT *, pg_terminate_backend(${pid}) FROM q WHERE pid <> pg_backend_pid()`
)
