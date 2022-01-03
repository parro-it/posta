package main

import (
	"log"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const (
	COLUMN_VERSION = iota
	COLUMN_FEATURE
)

func main() {
	// Initialize GTK without parsing any command line arguments.
	gtk.Init(nil)

	// Create a new toplevel window, set its title, and connect it to the
	// "destroy" signal to exit the GTK main loop when it is destroyed.
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.SetTitle("Simple Example")
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	// Create a new label widget to show in the window.
	l, err := gtk.LabelNew("Hello, gotk3!")
	if err != nil {
		log.Fatal("Unable to create label:", err)
	}

	store, err := gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING)
	if err != nil {
		log.Fatal("Unable to create store:", err)
	}
	folders, err := gtk.TreeViewNewWithModel(store)
	if err != nil {
		log.Fatal("Unable to create treeview:", err)
	}

	renderer, err := gtk.CellRendererTextNew()
	if err != nil {
		log.Fatal(err)
	}
	renderer.Set("editable", true)
	renderer.Set("editable-set", true)
	renderer.Connect("edited", func(_ *gtk.CellRendererText, path, text string) {
		iter, err := store.GetIterFromString(path)
		if err == nil {
			store.Set(iter, []int{0}, []interface{}{text})
		}
	})

	col, err := gtk.TreeViewColumnNewWithAttribute("Label", renderer,
		"text", 0)
	if err != nil {
		log.Fatal(err)
	}
	col.SetExpand(true)
	folders.AppendColumn(col)
	cr, err := gtk.CellRendererTextNew()
	if err != nil {
		log.Fatal(err)
	}
	col, err = gtk.TreeViewColumnNewWithAttribute("Address", cr, "text", 1)
	if err != nil {
		log.Fatal(err)
	}
	col.SetMinWidth(350)
	folders.AppendColumn(col)

	// Get an iterator for a new row at the end of the list store
	iter := store.Append()

	// Set the contents of the list store row that the iterator represents
	err = store.Set(iter,
		[]int{COLUMN_VERSION, COLUMN_FEATURE},
		[]interface{}{"version", "feature"})

	if err != nil {
		log.Fatal("Unable to add row:", err)
	}

	_ = l
	// Add the label to the window.
	win.Add(folders)

	// Set the default window size.
	win.SetDefaultSize(800, 600)

	// Recursively show all widgets contained in this window.
	win.ShowAll()

	// Begin executing the GTK main loop.  This blocks until
	// gtk.MainQuit() is run.
	gtk.Main()
}
