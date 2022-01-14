package main

import (
	"context"
	"log"

	"github.com/gotk3/gotk3/gtk"
	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/folders"
	"github.com/parro-it/posta/imap"
	"github.com/parro-it/posta/msgs"
)

const (
	COLUMN_VERSION = iota
	COLUMN_FEATURE
)

// Create and initialize the window
func mainWindow() *gtk.Window {
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}

	win.SetTitle("Posta")
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	win.SetPosition(gtk.WIN_POS_CENTER)
	win.SetDefaultSize(800, 600)

	horzCnt, err := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	if err != nil {
		log.Fatal(err)
	}

	horzCnt.Pack1(folders.View(), true, true)
	horzCnt.Pack2(msgs.View(), true, true)
	horzCnt.SetPosition(250)
	win.Add(horzCnt)

	return win
}

func main() {

	app.Instance.Start(context.Background(),
		folders.Start,
		imap.Start,
		msgs.Start,
	)

	gtk.Init(nil)

	win := mainWindow()

	/*
		selection, err := treeView.GetSelection()
		if err != nil {
			log.Fatal("Could not get tree selection object.")
		}
		selection.SetMode(gtk.SELECTION_SINGLE)
		selection.Connect("changed", treeSelectionChangedCB)
	*/

	win.ShowAll()
	gtk.Main()
}
