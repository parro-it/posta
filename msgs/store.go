package msgs

import (
	"log"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/parro-it/posta/app"
)

func NewStore() *gtk.TreeStore {
	store, err := gtk.TreeStoreNew(glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING)
	if err != nil {
		log.Fatal("Unable to create tree store:", err)
	}
	ch := app.ListenAction2[AddMsg, ClearMsgs]()

	go func() {

		//folders := map[string]*gtk.TreeIter{}

		for a := range ch {
			a := a
			glib.IdleAdd(func() bool {
				handleActions(a, store)
				return false
			})
		}
	}()

	return store
}

func handleActions(a any, store *gtk.TreeStore) {
	switch a := a.(type) {
	case ClearMsgs:
		store.Clear()
	case AddMsg:
		m := a.Msg
		msg := store.Append(nil)

		if err := store.SetValue(msg, COLUMN_SUBJECT, m.Subject); err != nil {
			log.Fatal("Unable set value:", err)
		}

		if err := store.SetValue(msg, COLUMN_DATE, m.Date.Format("2006-01-02 15:04")); err != nil {
			log.Fatal("Unable set value:", err)
		}

		if err := store.SetValue(msg, COLUMN_FROM, m.From); err != nil {
			log.Fatal("Unable set value:", err)
		}
		/*
			if err := store.SetValue(msg, COLUMN_TO, m.To); err != nil {
				log.Fatal("Unable set value:", err)
			}
		*/
	}
}
