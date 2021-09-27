package main

import (
	"context"
	"time"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-server"
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
	return ErrNotImplemented
	/*
		return this.Do(func(txn SQTransaction) error {
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
	*/
}

func (this *plugin) indexEvent(ctx context.Context, e Event) (int64, error) {
	return 0, ErrNotImplemented
	/*
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
		}*/
}
