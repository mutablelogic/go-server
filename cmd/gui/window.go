package main

import (
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/objc"
)

type Window struct {
	appkit.Window
}

func NewWindow(title string, width, height float64, view appkit.IView) *Window {
	w := appkit.NewWindowWithSize(width, height)
	objc.Retain(&w)
	w.SetTitle(title)
	w.ContentView().AddSubview(view)
	return &Window{w}
}

func (w *Window) Show() *Window {
	w.MakeKeyAndOrderFront(nil)
	w.Center()
	return w
}
