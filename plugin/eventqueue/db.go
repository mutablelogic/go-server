package main

import (
	"context"
	"fmt"
	"time"

	// Modules
	. "github.com/djthorpe/go-server"
	provider "github.com/djthorpe/go-server/pkg/provider"
	sq "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
	sqobj "github.com/djthorpe/go-sqlite/pkg/sqobj"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type SQEvent struct {
	Timestamp time.Time `sqlite:"timestamp,notnull"`
	Plugin    string    `sqlite:"plugin,notnull,index:plugin"`
	Name      string    `sqlite:"name,notnull,index:name"`
	Value     string    `sqlite:"value"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	tableNameEvent = "event"
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *plugin) createTables(context.Context) error {
	return this.Do(func(txn sq.SQTransaction) error {
		source := N(tableNameEvent).WithSchema(this.schema)
		for _, statement := range sqobj.CreateTableAndIndexes(source, true, SQEvent{}) {
			fmt.Println(statement)
			if _, err := txn.Exec(statement); err != nil {
				return err
			}
		}
		// Return success
		return nil
	})
}

func (this *plugin) indexEvent(ctx context.Context, e Event) (int64, error) {
	// Insert import configuration into database
	if params, err := sqobj.InsertParams(SQEvent{
		Timestamp: time.Now(),
		Plugin:    provider.ContextPluginName(ctx),
		Name:      e.Name(),
		Value:     fmt.Sprint(e.Value()),
	}); err != nil {
		return 0, err
	} else if r, err := this.Exec(sqobj.InsertRow(tableNameEvent, SQEvent{}), params...); err != nil {
		return 0, err
	} else if r.RowsAffected != 1 {
		return 0, ErrInternalAppError
	} else {
		return r.LastInsertId, nil
	}
}
