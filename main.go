package main

import (
	"context"
	"log"

	"github.com/gotk3/gotk3/gtk"
	"github.com/parro-it/posta/actions"
	"github.com/parro-it/posta/config"
	"github.com/parro-it/posta/folders"
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
	config.ParseCommandLine()
	err := config.Init()
	if err != nil {
		log.Fatal(err)
	}
	go actions.Start()
	gtk.Init(nil)

	win := mainWindow()

	win.Add(folders.View())
	go func() {
		err := <-folders.ReadFolders(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		/*
			err = <-folders.ListenUpdates(context.Background())
			if err != nil {
				log.Fatal(err)
			}
		*/
	}()

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
