package msgs

import (
	"log"

	"github.com/gotk3/gotk3/gtk"
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

const (
	COLUMN_DATE = iota
	COLUMN_FROM
	//COLUMN_TO
	COLUMN_SUBJECT
)

// Creates a tree view and the tree store that holds its data
func View() *gtk.ScrolledWindow {
	tree, err := gtk.TreeViewNew()
	if err != nil {
		log.Fatal("Unable to create tree view:", err)
	}
	tree.AppendColumn(createTextColumn("Date", COLUMN_DATE))
	tree.AppendColumn(createTextColumn("From", COLUMN_FROM))
	//tree.AppendColumn(createTextColumn("To", COLUMN_TO))
	tree.AppendColumn(createTextColumn("Subject", COLUMN_SUBJECT))

	// Creating a tree store. This is what holds the data that will be shown on our tree view.
	store := NewStore()
	tree.SetModel(store)

	scroll, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		log.Fatal("Unable to create scroolbox for tree view:", err)
	}
	scroll.Add(tree)
	return scroll
}

/*
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
*/
