package folders

import (
	"fmt"
	"log"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/parro-it/posta/app"
)

func NewStore() *gtk.TreeStore {
	store, err := gtk.TreeStoreNew(glib.TYPE_OBJECT, glib.TYPE_STRING)
	if err != nil {
		log.Fatal("Unable to create tree store:", err)
	}

	go func() {
		ch := app.ListenAction[Added]()
		folders := map[string]*gtk.TreeIter{}

		for a := range ch {
			a := a
			glib.IdleAdd(func() bool {
				handleActions(a, store, folders)
				return false
			})
		}
	}()

	return store
}

func handleActions(a any, store *gtk.TreeStore, folders map[string]*gtk.TreeIter) {
	switch a := a.(type) {
	case Added:
		var parent *gtk.TreeIter
		var account *gtk.TreeIter
		var ok bool

		if account, ok = folders[a.Folder.Account]; !ok {
			account = store.Append(nil)
			folders[a.Folder.Account] = account
			err := store.SetValue(account, COLUMN_ICON, imageFolder)
			if err != nil {
				log.Fatal("Unable set value:", err)
			}
			err = store.SetValue(account, COLUMN_TEXT, a.Folder.Account)
			if err != nil {
				log.Fatal("Unable set value:", err)
			}
		}

		if len(a.Folder.Path) > 1 {
			parentPath := a.Folder.Path[0 : len(a.Folder.Path)-1]
			if parent, ok = folders[strings.Join(parentPath, "/")]; !ok {
				fmt.Printf("Parent not found for %v\n", parentPath)
				return
			}
		} else {
			parent = account
		}

		i := store.Append(parent)
		folders[strings.Join(a.Folder.Path, "/")] = i
		// Set the contents of the tree store row that the iterator represents
		err := store.SetValue(i, COLUMN_ICON, imageFolder)
		if err != nil {
			log.Fatal("Unable set value:", err)
		}
		fmt.Printf("set folder in store %s\n", a.Folder.Name)
		err = store.SetValue(i, COLUMN_TEXT, a.Folder.Name)
		if err != nil {
			log.Fatal("Unable set value:", err)
		}

	}
}

// Icons Pixbuf representation
var (
	imageFolder *gdk.Pixbuf = nil
)

// Load the icon image data from file:
func init() {
	iconPath := "/mnt/repos/parro-it/posta/"

	var err error
	imageFolder, err = gdk.PixbufNewFromFile(fmt.Sprintf("%s/folder.png", iconPath))
	if err != nil {
		log.Fatal("Unable to load image:", err)
	}
}

// IDs to access the tree view columns by
const (
	COLUMN_ICON = iota
	COLUMN_TEXT
)
