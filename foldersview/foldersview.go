package main

import (
	"context"
	"log"

	"github.com/gotk3/gotk3/gtk"
	"github.com/parro-it/posta/actions"
	"github.com/parro-it/posta/config"
	"github.com/parro-it/posta/folders"
)

// IDs to access the tree view columns by
const (
	COLUMN_ICON = iota
	COLUMN_TEXT
)

// Add a column to the tree view (during the initialization of the tree view)
// We need to distinct the type of data shown in either column.
func createTextColumn(title string, id int) *gtk.TreeViewColumn {
	// In this column we want to show text, hence create a text renderer
	cellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		log.Fatal("Unable to create text cell renderer:", err)
	}

	// Tell the renderer where to pick input from. Text renderer understands
	// the "text" property.
	column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "text", id)
	if err != nil {
		log.Fatal("Unable to create cell column:", err)
	}

	return column
}

func createImageColumn(title string, id int) *gtk.TreeViewColumn {
	// In this column we want to show image data from Pixbuf, hence
	// create a pixbuf renderer
	cellRenderer, err := gtk.CellRendererPixbufNew()
	if err != nil {
		log.Fatal("Unable to create pixbuf cell renderer:", err)
	}

	// Tell the renderer where to pick input from. Pixbuf renderer understands
	// the "pixbuf" property.
	column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "pixbuf", id)
	if err != nil {
		log.Fatal("Unable to create cell column:", err)
	}

	return column
}

// Creates a tree view and the tree store that holds its data
func setupTreeView() *gtk.ScrolledWindow {
	tree, err := gtk.TreeViewNew()
	if err != nil {
		log.Fatal("Unable to create tree view:", err)
	}

	tree.AppendColumn(createImageColumn("Icon", COLUMN_ICON))
	tree.AppendColumn(createTextColumn("Version", COLUMN_TEXT))

	// Creating a tree store. This is what holds the data that will be shown on our tree view.
	store := folders.NewStore()
	tree.SetModel(store)

	scroll, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		log.Fatal("Unable to create scroolbox for tree view:", err)
	}
	scroll.Add(tree)
	return scroll
}

// Create and initialize the window
func setupWindow(title string) *gtk.Window {
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}

	win.SetTitle(title)
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	win.SetPosition(gtk.WIN_POS_CENTER)
	width, height := 600, 300
	win.SetDefaultSize(width, height)
	return win
}

// Handle selection
func treeSelectionChangedCB(selection *gtk.TreeSelection) {
	var iter *gtk.TreeIter
	var model gtk.ITreeModel
	var ok bool
	model, iter, ok = selection.GetSelected()
	if ok {
		tpath, err := model.(*gtk.TreeModel).GetPath(iter)
		if err != nil {
			log.Printf("treeSelectionChangedCB: Could not get path from model: %s\n", err)
			return
		}
		log.Printf("treeSelectionChangedCB: selected path: %s\n", tpath)
	}
}

func main() {
	config.ParseCommandLine()
	err := config.Init()
	if err != nil {
		log.Fatal(err)
	}
	go actions.Start()
	gtk.Init(nil)

	win := setupWindow("Go Example Testreport")

	//var iter1, iter2 *gtk.TreeIter

	treeView := setupTreeView()
	win.Add(treeView)
	go func() {
		err := <-folders.ReadFolders(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}()

	/*

		go func() {
			f := folders.ReadFolders(context.Background())
			for f := range f.Folders {
				fmt.Printf("post aciot for %s\n", f.Name)
				actions.Post(folders.Added{Folder: f})
			}
			if f.Err != nil {
				log.Fatal(f.Err)
			}
		}()
			// Add some rows to the tree store
			iter1 = addRow(treeStore, imageOK, "Testsuite 1")
			iter2 = addSubRow(treeStore, iter1, imageOK, "test1-1")
			iter2 = addSubRow(treeStore, iter1, imageOK, "test1-2")
			addSubRow(treeStore, iter2, imageOK, "test1-2-1")
			addSubRow(treeStore, iter2, imageOK, "test1-2-2")
			addSubRow(treeStore, iter2, imageOK, "test1-2-3")
			iter2 = addSubRow(treeStore, iter1, imageOK, "test1-3")
			iter1 = addRow(treeStore, imageFAIL, "Testsuite 2")
			iter2 = addSubRow(treeStore, iter1, imageOK, "test2-1")
			iter2 = addSubRow(treeStore, iter1, imageOK, "test2-2")
			iter2 = addSubRow(treeStore, iter1, imageFAIL, "test2-3")
			addSubRow(treeStore, iter2, imageOK, "test2-3-1")
			addSubRow(treeStore, iter2, imageFAIL, "test2-3-2")
	*/

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
