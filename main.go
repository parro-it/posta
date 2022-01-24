package main

import (
	"context"
	"log"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/folders"
	"github.com/parro-it/posta/imap"
	"github.com/parro-it/posta/msgbody"
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
	win.SetDefaultSize(1600, 800)

	horzCnt1, err := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	if err != nil {
		log.Fatal(err)
	}

	horzCnt1.Pack1(folders.View(), false, false)
	horzCnt2, err := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	if err != nil {
		log.Fatal(err)
	}
	horzCnt2.Pack1(msgs.View(), true, true)
	horzCnt2.Pack2(msgbody.View(), false, false)
	horzCnt1.Pack2(horzCnt2, true, true)

	horzCnt1.SetPosition(200)
	horzCnt2.SetPosition(1000)
	horzCnt1.SetWideHandle(true)
	win.Add(horzCnt1)
	win.Connect("key-press-event", func(win *gtk.Window, ev *gdk.Event) {
		keyEvent := &gdk.EventKey{Event: ev}

		app.Instance.PostKeyPressed(keyEvent.KeyVal(), keyEvent.State())

	})
	return win
}

func main() {

	app.Instance.Start(context.Background(),
		folders.Start,
		imap.Start,
		msgs.Start,
		msgbody.Start,
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
