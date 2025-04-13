package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/progrium/darwinkit/dispatch"
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"
	"github.com/progrium/darwinkit/objc"
)

type DataSource struct {
	values        chan [][2]objc.Object
	data          [][2]objc.Object
	delegate      *TableViewDataSourceDelegate
	wg            *sync.WaitGroup
	maxRows       int
	overflowValue int
}

func NewDataSource(ctx context.Context, maxRows int, overflowValue int) *DataSource {
	model := &DataSource{
		maxRows:       maxRows,
		overflowValue: overflowValue,
		values:        make(chan [][2]objc.Object, 1),
		data:          nil,
		delegate:      &TableViewDataSourceDelegate{},
		wg:            &sync.WaitGroup{},
	}
	model.delegate.SetTableViewObjectValueForTableColumnRow(func(tableView appkit.TableView, tableColumn appkit.TableColumn, row int) objc.Object {
		model.getLatest(ctx)
		if model.data == nil {
			return objc.ObjectFrom(nil)
		}
		switch tableColumn.Identifier() {
		case "Column1":
			return model.data[row][0]
		case "Column2":
			return model.data[row][1]
		}
		panic("unknown column")
	})
	model.delegate.SetNumberOfRowsInTableView(func(tableView appkit.TableView) int {
		model.getLatest(ctx)
		return len(model.data)
	})
	return model
}

func (t *DataSource) getLatest(ctx context.Context) {
	switch t.data {
	case nil:
		select {
		case t.data = <-t.values:
		case <-ctx.Done():
			t.data = nil
			return
		}
	default:
		select {
		case t.data = <-t.values:
		case <-ctx.Done():
			t.data = nil
			return
		default:
		}
	}
}

func (t *DataSource) Wait() {
	t.wg.Wait()
}

func (t *DataSource) Start(ctx context.Context, tableView appkit.TableView) {
	t.wg.Add(1)
	go valueGenerator(ctx, t.wg, t.values, tableView, t.maxRows, t.overflowValue)
}

func valueGenerator(ctx context.Context, wg *sync.WaitGroup, values chan<- [][2]objc.Object, tableView appkit.TableView, maxRows int, overflowValue int) {
	defer wg.Done()
	currentValues := make([][2]int, maxRows)
	sendToUI := func() {
		snapshot := make([][2]objc.Object, len(currentValues))
		for i, v := range currentValues {
			snapshot[i] = [2]objc.Object{foundation.String_StringWithString(fmt.Sprintf("%d", v[0])).Object, foundation.String_StringWithString(fmt.Sprintf("%d", v[1])).Object}
		}
		select {
		case <-ctx.Done():
			return
		case values <- snapshot:
			dispatch.MainQueue().DispatchAsync(func() {
				tableView.SetNeedsDisplay(true)
			})
		}
	}
	counter := 0
	row := 0
	for {
		row = counter / 2
		if row >= maxRows {
			row = 0
			break
		}
		currentValues[row][counter%2] = counter % overflowValue
		counter++
	}
	sendToUI()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(10 * time.Millisecond)
			row = (counter / 2) % maxRows
			currentValues[row][counter%2] = counter % overflowValue
			counter++
			sendToUI()
		}
	}

}
