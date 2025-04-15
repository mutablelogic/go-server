package schema

import (
	"context"
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

///////////////////////////////////////////////////////////////////////////////////
// WRITER

func (user UserMeta) Insert(bind *pg.Bind) (string, error) {
	bind.Set("name", types.TrimStringPtr(user.Name))
	bind.Set("desc", types.TrimStringPtr(user.Desc))
	if scope := user.Scope; scope == nil {
		bind.Set("scope", "{}")
	} else {
		bind.Set("scope", scope)
	}
	if meta := user.Meta; meta == nil {
		bind.Set("meta", "{}")
	} else {
		bind.Set("meta", user.Meta)
	}
	return userUpsert, nil
}

func (user UserMeta) Update(bind *pg.Bind) error {
	bind.Del("patch")
	if user.Name != nil {
		bind.Append("patch", `"name" = `+bind.Set("name", types.TrimStringPtr(user.Name)))
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

///////////////////////////////////////////////////////////////////////////////////
// READER

func (user *User) Scan(row pg.Row) error {
	return row.Scan(&user.Name, &user.Ts, &user.Status, &user.Desc, &user.Scope, &user.Meta)
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
	userDelete = `
		DELETE FROM
			${"schema"}."user"
		WHERE 
			name = @id
		RETURNING
			"name", "ts", "status", "desc", "scope", "meta"
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
)
