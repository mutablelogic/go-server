package main

import (
	"context"

	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"
)

type TableView struct {
	appkit.TableView
}

const (
	RowHeight = 20
)

func NewTableView(ctx context.Context, dataSource *DataSource) *TableView {
	tableView := new(TableView)
	tableView.TableView = appkit.NewTableView()
	go dataSource.Start(ctx, tableView.TableView)

	tableView.SetRowHeight(RowHeight)
	tableView.SetTranslatesAutoresizingMaskIntoConstraints(false)
	tableView.SetHeaderView(appkit.NewTableHeaderViewWithFrame(rectOf(0, 0, 0, RowHeight)))
	tableView.SetGridStyleMask(appkit.TableViewSolidVerticalGridLineMask | appkit.TableViewSolidHorizontalGridLineMask)
	tableView.SetStyle(appkit.TableViewStylePlain)
	tableView.SetRowSizeStyle(appkit.TableViewRowSizeStyleDefault)
	tableView.SetColumnAutoresizingStyle(appkit.TableViewUniformColumnAutoresizingStyle)
	tableView.SetUsesAlternatingRowBackgroundColors(true)
	tableView.SetStyle(appkit.TableViewStyleFullWidth)
	tableView.SetTranslatesAutoresizingMaskIntoConstraints(false)
	tableColumn1 := appkit.NewTableColumn().InitWithIdentifier("Column1")
	tableColumn1.SetTitle("Test 1")
	tableColumn1.SetWidth(100)
	tableView.AddTableColumn(tableColumn1)
	tableColumn2 := appkit.NewTableColumn().InitWithIdentifier("Column2")
	tableColumn2.SetTitle("Test 2")
	tableColumn2.SetWidth(100)
	tableView.AddTableColumn(tableColumn2)
	tableView.SetDataSource(dataSource.delegate)
	tableView.SetAllowsColumnSelection(true)
	tableView.SetAutoresizingMask(appkit.ViewWidthSizable | appkit.ViewHeightSizable)
	return tableView
}

func rectOf(x, y, width, height float64) foundation.Rect {
	return foundation.Rect{Origin: foundation.Point{X: x, Y: y}, Size: foundation.Size{Width: width, Height: height}}
}
