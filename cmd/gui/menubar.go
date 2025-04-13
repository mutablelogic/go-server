package main

import (
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/objc"
)

type MenuBar struct {
	appkit.Menu
	items []appkit.IMenuItem
}

func NewMenuBar(title string) *MenuBar {
	menubar := new(MenuBar)
	menubar.Menu = appkit.NewMenuWithTitle(title)

	appMenuItem := appkit.NewMenuItemWithSelector(title, "", objc.Selector{})
	appMenu := appkit.NewMenuWithTitle(title)
	//	appMenu.AddItem(appkit.NewMenuItemWithAction("Quit", "q", func(sender objc.Object) {}))
	appMenuItem.SetSubmenu(appMenu)
	menubar.AddItem(appMenuItem)

	return menubar
}

/*
func do() {
	menuBar := appkit.NewMenuWithTitle("Table View")
	a.app.SetMainMenu(menuBar)

	appMenuItem := appkit.NewMenuItemWithSelector("Table View", "", objc.Selector{})
	appMenu := appkit.NewMenuWithTitle("Table View")
	appMenu.AddItem(appkit.NewMenuItemWithAction("Quit", "q", func(sender objc.Object) { a.app.Terminate(nil) }))
	appMenuItem.SetSubmenu(appMenu)
	menuBar.AddItem(appMenuItem)
}
*/
