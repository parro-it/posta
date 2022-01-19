package folders

import (
	"fmt"
	"log"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/parro-it/posta/app"
	imapProc "github.com/parro-it/posta/imap"
)

var folders = map[string]*gtk.TreeIter{}
var foldersObj = map[string]*imapProc.Folder{}

func NewStore() *gtk.TreeStore {
	store, err := gtk.TreeStoreNew(glib.TYPE_OBJECT, glib.TYPE_STRING, glib.TYPE_STRING)
	if err != nil {
		log.Fatal("Unable to create tree store:", err)
	}
	ch := app.ListenAction[Added]()

	go func() {
		store.SetProperty("map", folders)
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

		if strings.Contains(a.Folder.Path, a.Folder.Sep) {
			path := strings.Split(a.Folder.Path, a.Folder.Sep)
			parentPath := path[0 : len(path)-1]
			if parent, ok = folders[strings.Join(parentPath, a.Folder.Sep)]; !ok {
				fmt.Printf("Parent not found for %v\n", parentPath)
				return
			}
		} else {
			parent = account
		}

		i := store.Append(parent)
		folders[a.Folder.Path] = i
		foldersObj[a.Folder.Path] = &a.Folder

		// Set the contents of the tree store row that the iterator represents
		err := store.SetValue(i, COLUMN_ICON, imageFolder)
		if err != nil {
			log.Fatal("Unable set value:", err)
		}
		err = store.SetValue(i, COLUMN_TEXT, a.Folder.Name)
		if err != nil {
			log.Fatal("Unable set value:", err)
		}

		err = store.SetValue(i, COLUMN_OBJ, a.Folder.Path)
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
	COLUMN_OBJ
)
