package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	// Modules
	driver "github.com/djthorpe/go-sqlite"
	sqlite "github.com/djthorpe/go-sqlite/pkg/sqlite"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type sq struct {
	Timezone string                 `yaml:"timezone"`
	Database map[string]interface{} `yaml:"databases"`

	driver.SQConnection
	tz *time.Location
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	tickerDelta = time.Second * 5
)

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the module
func New(ctx context.Context, provider Provider) Plugin {
	this := new(sq)

	// Load configuration
	if err := provider.GetConfig(ctx, this); err != nil {
		provider.Print(ctx, "GetConfig: ", err)
		return nil
	}

	// Lookup timezone
	if tz, err := time.LoadLocation(this.Timezone); err != nil {
		provider.Print(ctx, "Timezone: ", err)
		return nil
	} else {
		this.tz = tz
	}

	// Load databases
	if db, err := this.OpenDatabases(ctx, provider, this.Database); err != nil {
		provider.Print(ctx, "OpenDatabases: ", err)
		return nil
	} else {
		this.SQConnection = db
	}

	// Add handler for ping
	if err := provider.AddHandlerFuncEx(ctx, reRoutePing, this.ServePing); err != nil {
		provider.Print(ctx, "Failed to add handler: ", err)
		return nil
	}

	// Add handler for schema
	if err := provider.AddHandlerFuncEx(ctx, reRouteSchema, this.ServeSchema); err != nil {
		provider.Print(ctx, "Failed to add handler: ", err)
		return nil
	}

	// Add handler for importing data
	if err := provider.AddHandlerFuncEx(ctx, reRouteImport, this.ServeImport, http.MethodPost); err != nil {
		provider.Print(ctx, "Failed to add handler: ", err)
		return nil
	}

	// Add handler for table or view
	if err := provider.AddHandlerFuncEx(ctx, reRouteTable, this.ServeTable); err != nil {
		provider.Print(ctx, "Failed to add handler: ", err)
		return nil
	}

	// Return success
	return this
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Name() string {
	return "sqlite"
}

func Version() string {
	return sqlite.Version()
}

// Open in-memory database (with no arguments) or several databases
// with each database being attached as a different schema
func (this *sq) OpenDatabases(ctx context.Context, provider Provider, attach map[string]interface{}) (driver.SQConnection, error) {
	// If no databases listed, then open in-memory database
	if len(attach) == 0 {
		return sqlite.New(this.tz)
	}

	// Otherwise, if there is no database called "main"
	// then open a database called "main"
	var db driver.SQConnection
	if path, exists := attach["main"]; exists {
		if path, ok := path.(string); ok && path != "" {
			provider.Printf(ctx, "Attaching %q: %q", "main", path)
			if db_, err := sqlite.Open(path, this.tz); err != nil {
				return nil, err
			} else {
				db = db_
			}
		}
	} else {
		if db_, err := sqlite.New(this.tz); err != nil {
			return nil, err
		} else {
			db = db_
		}
	}
	// Attach all other databases
	for schema, path := range attach {
		if schema == "main" {
			continue
		}
		if path, ok := path.(string); ok && path != "" {
			provider.Printf(ctx, "Attaching %q: %q", schema, path)
			if err := db.Attach(schema, path); err != nil {
				db.Close()
				return nil, err
			}
		}
	}
	// Return database connection
	return db, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *sq) String() string {
	str := "<sqlite"
	str += " version=" + strconv.Quote(Version())
	if this.SQConnection != nil {
		if modules := this.SQConnection.Modules(); modules != nil {
			str += fmt.Sprintf(" modules=%q", modules)
		}
		if schemas := this.SQConnection.Schemas(); schemas != nil {
			str += fmt.Sprintf(" schemas=%q", schemas)
		}
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *sq) Run(ctx context.Context, provider Provider) error {
	var result error

	// Create all the tables needed for importing data in the background
	if err := this.CreateImportTables("main"); err != nil {
		return err
	}

	// Create a ticker to periodically check for new import jobs
	ticker := time.NewTicker(tickerDelta)
	defer ticker.Stop()

FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		case <-ticker.C:
			if err := this.GetImportJob(); err != nil {
				provider.Print(ctx, "GetImportJob: ", err)
			}
		}
	}

	// Close database connection, release resources
	if this.SQConnection != nil {
		result = this.SQConnection.Close()
		this.SQConnection = nil
	}

	// Return any errors
	return result
}
