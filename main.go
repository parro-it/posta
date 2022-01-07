package main

import (
	"context"
	"log"

	"github.com/gotk3/gotk3/gtk"
	"github.com/parro-it/posta/actions"
	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/folders"
	"github.com/parro-it/posta/login"
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
	return win
}

func main() {
	go actions.Start()

	go func() {
		errs := folders.Start(context.Background())

		errs2 := login.Start(context.Background())
		errs3 := msgs.Start(context.Background())

		for {
			select {
			case e := <-errs:
				if e != nil {
					panic(e)
				}
			case e := <-errs2:
				if e != nil {
					panic(e)
				}

			case e := <-errs3:
				if e != nil {
					panic(e)
				}
			}
		}
		/*
			err = <-folders.ListenUpdates(context.Background())
			if err != nil {
				log.Fatal(err)
			}
		*/
	}()

	go app.Instance.Start()
	gtk.Init(nil)

	win := mainWindow()

	win.Add(folders.View())

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
