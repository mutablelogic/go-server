package schema

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	// Packages
	pg "github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////////
// TYPES

type UserName string

type UserStatus string

type UserTokenHash string

type UserMeta struct {
	Name  *string        `json:"name,omitempty" arg:"" help:"Name"`
	Desc  *string        `json:"desc,omitempty" help:"Description"`
	Scope []string       `json:"scope,omitempty" help:"Scopes"`
	Meta  map[string]any `json:"meta,omitempty" help:"Metadata"`
}

type User struct {
	UserMeta
	Status string    `json:"status,omitempty" help:"Status"`
	Ts     time.Time `json:"ts,omitempty" help:"Timestamp"`
}

type UserListRequest struct {
	Status *string `json:"status,omitempty" help:"Status"`
	Scope  *string `json:"scope,omitempty" help:"Scope"`
	pg.OffsetLimit
}

type UserList struct {
	UserListRequest
	Count uint64 `json:"count"`
	Body  []User `json:"body,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////////
// TYPES

const (
	UserStatusLive     = UserStatus("live")
	UserStatusArchived = UserStatus("archived")
)

///////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (user UserMeta) String() string {
	data, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (user User) String() string {
	data, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (user UserListRequest) String() string {
	data, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (user UserList) String() string {
	data, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////////
// SELECTOR

func (user UserName) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Name
	if name := strings.TrimSpace(string(user)); name == "" {
		return "", httpresponse.ErrBadRequest.With("name is missing")
	} else {
		bind.Set("id", name)
	}

	switch op {
	case pg.Get:
		return userGet, nil
	case pg.Delete:
		return userDelete, nil
	case pg.Update:
		return userUpdate, nil
	default:
		return "", httpresponse.ErrBadRequest.Withf("UserName: operation %q is not supported", op)
	}
}

func (hash UserTokenHash) Select(bind *pg.Bind, op pg.Op) (string, error) {
	bind.Set("hash", hash)

	switch op {
	case pg.Get:
		return userFromTokenHash, nil
	default:
		return "", httpresponse.ErrBadRequest.Withf("UserName: operation %q is not supported", op)
	}
}

func (list *UserListRequest) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Order
	bind.Set("orderby", `ORDER BY ts DESC`)

	// Where
	bind.Del("where")

	// Status
	if status := types.TrimStringPtr(list.Status); status == nil {
		list.Status = types.StringPtr(string(UserStatusLive))
	} else {
		list.Status = status
	}
	if list.Status != nil {
		bind.Append("where", `status = `+bind.Set("status", list.Status))
	}

	// Scope
	list.Scope = types.TrimStringPtr(list.Scope)
	if list.Scope != nil {
		bind.Append("where", `scope @> `+bind.Set("scope", []string{types.PtrString(list.Scope)}))
	}

	// Where AND
	if where := bind.Join("where", " AND "); where != "" {
		bind.Set("where", `WHERE `+where)
	} else {
		bind.Set("where", "")
	}

	// Offset and limit
	list.OffsetLimit.Bind(bind, UserListLimit)

	switch op {
	case pg.List:
		return userList, nil
	default:
		return "", httpresponse.ErrBadRequest.Withf("UserListRequest: operation %q is not supported", op)
	}
}

///////////////////////////////////////////////////////////////////////////////////
// WRITER

func (user UserMeta) Insert(bind *pg.Bind) (string, error) {
	// Name cannot include path characters
	if name, err := UserName(types.PtrString(user.Name)).validate(); err != nil {
		return "", err
	} else {
		bind.Set("name", name)
	}

	// Set description
	bind.Set("desc", types.TrimStringPtr(user.Desc))

	// Set scope
	if scope := user.Scope; scope == nil {
		bind.Set("scope", "{}")
	} else {
		bind.Set("scope", scope)
	}

	// Set meta
	if meta := user.Meta; meta == nil {
		bind.Set("meta", "{}")
	} else {
		bind.Set("meta", user.Meta)
	}

	// Return success
	return userUpsert, nil
}

func (user UserMeta) Update(bind *pg.Bind) error {
	bind.Del("patch")
	if user.Name != nil {
		if name, err := UserName(types.PtrString(user.Name)).validate(); err != nil {
			return err
		} else {
			bind.Append("patch", `"name" = `+bind.Set("name", name))
		}
	}
	if user.Desc != nil {
		bind.Append("patch", `"desc" = `+bind.Set("desc", types.TrimStringPtr(user.Desc)))
	}
	if user.Scope != nil {
		bind.Append("patch", `"scope" = `+bind.Set("scope", user.Scope))
	}
	if user.Meta != nil {
		bind.Append("patch", `"meta" = `+bind.Set("meta", user.Meta))
	}

	// Set patch
	if patch := bind.Join("patch", ", "); patch != "" {
		bind.Set("patch", patch)
	} else {
		return httpresponse.ErrBadRequest.With("no fields to update")
	}

	// Return success
	return nil
}

func (status UserStatus) Insert(bind *pg.Bind) (string, error) {
	return "", httpresponse.ErrNotImplemented.With("UserStatus.Insert")
}

func (status UserStatus) Update(bind *pg.Bind) error {
	if status == UserStatusLive || status == UserStatusArchived {
		bind.Set("patch", `status = `+bind.Set("status", status))
	} else {
		return httpresponse.ErrBadRequest.Withf("invalid status %q", status)
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////////
// READER

func (user *User) Scan(row pg.Row) error {
	return row.Scan(&user.Name, &user.Ts, &user.Status, &user.Desc, &user.Scope, &user.Meta)
}

func (list *UserList) ScanCount(row pg.Row) error {
	list.Body = []User{}
	return row.Scan(&list.Count)
}

func (list *UserList) Scan(row pg.Row) error {
	var user User
	if err := user.Scan(row); err != nil {
		return err
	}
	list.Body = append(list.Body, user)
	return nil
}

///////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (user UserName) validate() (string, error) {
	if name := strings.TrimSpace(string(user)); name == "" {
		return "", httpresponse.ErrBadRequest.With("name is missing")
	} else if strings.ContainsRune(name, os.PathSeparator) {
		return "", httpresponse.ErrBadRequest.With("name cannot contain path separator")
	} else {
		return name, nil
	}
}

///////////////////////////////////////////////////////////////////////////////////
// SQL

// Create objects in the schema
func bootstrapUser(ctx context.Context, conn pg.Conn) error {
	q := []string{
		userCreateStatusType,
		userCreateTable,
	}
	for _, query := range q {
		if err := conn.Exec(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

const (
	userCreateStatusType = `
		DO $$ BEGIN
			CREATE TYPE ${"schema"}.USER_STATUS AS ENUM ('live', 'archived');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;
	`
	userCreateTable = `
		CREATE TABLE IF NOT EXISTS ${"schema"}."user" (
			"name"             TEXT PRIMARY KEY,                                   -- unique name for the user			
			"ts"               TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP, -- timestamp of creation or update
			"status"           ${"schema"}.USER_STATUS NOT NULL DEFAULT 'live',    -- status of the user	
			"desc"             TEXT,                                               -- description of the user
			"scope"			   TEXT[] NOT NULL DEFAULT '{}',                       -- allowed scopes for the user
			"meta" 		       JSONB NOT NULL DEFAULT '{}'::JSONB                  -- additional metadata
		)	
	`
	userUpsert = `
		INSERT INTO ${"schema"}."user" (
			"name", "ts", "status", "desc", "scope", "meta"
		) VALUES (
		 	@name, DEFAULT, DEFAULT, @desc, @scope, @meta
		) ON CONFLICT ("name") DO UPDATE SET
			"ts" = CURRENT_TIMESTAMP,
			"desc" = @desc,
			"scope" = @scope,
			"meta" = @meta
		RETURNING
			"name", "ts", "status", "desc", "scope", "meta"
	`
	userSelect = `
		SELECT
			"name", "ts", "status", "desc", "scope", "meta"
		FROM
			${"schema"}."user"
	`
	userGet    = userSelect + ` WHERE name = @id`
	userList   = `WITH q AS (` + userSelect + `) SELECT * FROM q ${where} ${orderby}`
	userDelete = `
		DELETE FROM
			${"schema"}."user"
		WHERE 
			name = @id
		RETURNING
			"name", "ts", 'deleted' AS "status", "desc", "scope", "meta"
	`
	userUpdate = `
		UPDATE
			${"schema"}."user"
		SET
			${patch}, "ts" = CURRENT_TIMESTAMP
		WHERE
			name = @id
		RETURNING
			"name", "ts", "status", "desc", "scope", "meta"
	`
	userFromTokenHash = `
		SELECT
			U."name", U."ts", U."status", U."desc", U."scope", U."meta" || jsonb_build_object('token_id', T."id") AS "meta"
		FROM
			${"schema"}."user" U
		JOIN
			${"schema"}."token" T ON U."name" = T."user"
		WHERE
			T."hash" = @hash
		AND
			T."status" = 'live'	
	`
)
