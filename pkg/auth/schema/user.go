package schema

import (
	"context"

	// Packages
	"github.com/djthorpe/go-pg"
)

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
		CREATE TABLE IF NOT EXISTS ${"schema"}.user (
			"name"             TEXT PRIMARY KEY,                                   -- unique name for the user			
			"ts"               TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP, -- timestamp of creation or update
			"status"           ${"schema"}.USER_STATUS NOT NULL DEFAULT 'live',    -- status of the user	
			"desc"             TEXT,                                               -- description of the user
			"scope"			   TEXT[] NOT NULL DEFAULT '{}'                        -- allowed scopes for the user
			"meta" 		       JSONB NOT NULL DEFAULT '{}'::JSONB,                 -- additional metadata
		)	
	`
)
