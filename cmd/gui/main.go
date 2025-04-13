package main

import (
	"context"

	"github.com/progrium/darwinkit/helper/layout"
	"github.com/progrium/darwinkit/macos"
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"
)

func main() {
	macos.RunApp(func(app appkit.Application, delegate *appkit.ApplicationDelegate) {
		app.SetActivationPolicy(appkit.ApplicationActivationPolicyRegular)
		app.ActivateIgnoringOtherApps(true)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		dataSource := NewDataSource(ctx, 40, 13)
		tableView := NewTableView(ctx, dataSource)

		NewWindow("Test", 800, 600, tableView.TableView).Show()
		app.SetMainMenu(NewMenuBar("Test").Menu)

		layout.PinEdgesToSuperView(tableView, foundation.EdgeInsets{Top: 10, Bottom: 10, Left: 20, Right: 20})

		delegate.SetApplicationShouldTerminateAfterLastWindowClosed(func(appkit.Application) bool {
			return true
		})
	})
}
