package msgs

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/imap"
)

var dtFormat = "2006-01-02 15:04"

const (
	COLUMN_DATE = iota
	COLUMN_FROM
	COLUMN_SUBJECT
	COLUMN_OBJ
)

func NewStore() *gtk.TreeStore {
	store, err := gtk.TreeStoreNew(glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_INT)
	if err != nil {
		log.Fatal("Unable to create tree store:", err)
	}
	ch := app.ListenAction2[AddMsg, ClearMsgs]()
	store.SetSortFunc(COLUMN_DATE, func(model *gtk.TreeModel, a, b *gtk.TreeIter) int {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("Error sorting date: %s\n", err)
			}
		}()
		av, err := model.GetValue(a, COLUMN_DATE)
		if err != nil {
			//panic(err)
			return 0
		}
		var avs string
		avi, err := av.GoValue()
		if err != nil {
			//panic(err)
			return 0
		}
		avs = avi.(string)

		bv, err := model.GetValue(b, COLUMN_DATE)
		if err != nil {
			//panic(err)
			return 0
		}
		var bvs string
		bvi, err := bv.GoValue()
		if err != nil {
			//panic(err)
			return 0
		}
		bvs = bvi.(string)
		var adt time.Time
		var bdt time.Time

		if avs != "" {
			adt, err = time.Parse(dtFormat, avs)
			if err != nil {
				//panic(err)
				return 0
			}
		}

		if bvs != "" {
			bdt, err = time.Parse(dtFormat, bvs)
			if err != nil {
				//panic(err)
				return 0
			}
		}
		if adt == bdt {
			return 0
		}
		if adt.After(bdt) {
			return -1
		}
		return 1
	})

	store.SetSortFunc(COLUMN_FROM, func(model *gtk.TreeModel, a, b *gtk.TreeIter) int {

		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("Error sorting date: %s\n", err)
			}
		}()

		av, err := model.GetValue(a, COLUMN_FROM)
		if err != nil {
			//panic(err)
			return 0
		}
		var avs string
		avi, err := av.GoValue()
		if err != nil {
			//panic(err)
			return 0
		}
		avs = avi.(string)

		bv, err := model.GetValue(b, COLUMN_FROM)
		if err != nil {
			//panic(err)
			return 0
		}
		var bvs string
		bvi, err := bv.GoValue()
		if err != nil {
			//panic(err)
			return 0
		}
		bvs = bvi.(string)

		if avs == bvs {
			return 0
		}
		if avs > bvs {
			return -1
		}
		return 1
	})
	store.SetSortFunc(COLUMN_SUBJECT, func(model *gtk.TreeModel, a, b *gtk.TreeIter) int {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("Error sorting date: %s\n", err)
			}
		}()

		av, err := model.GetValue(a, COLUMN_SUBJECT)
		if err != nil {
			//panic(err)
			return 0
		}
		var avs string
		avi, err := av.GoValue()
		if err != nil {
			//panic(err)
			return 0
		}
		avs = avi.(string)

		bv, err := model.GetValue(b, COLUMN_SUBJECT)
		if err != nil {
			//panic(err)
			return 0
		}
		var bvs string
		bvi, err := bv.GoValue()
		if err != nil {
			//panic(err)
			return 0
		}
		bvs = bvi.(string)

		if avs == bvs {
			return 0
		}
		if avs > bvs {
			return -1
		}
		return 1
	})
	//store.SetSortColumnId(COLUMN_DATE, gtk.SORT_DESCENDING)
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

var mails = map[int]*imap.Msg{}

func handleActions(a any, store *gtk.TreeStore) {
	switch a := a.(type) {
	case ClearMsgs:
		mails = map[int]*imap.Msg{}
		store.Clear()
	case AddMsg:

		for _, m := range a.Msgs {

			msg := store.Append(nil)

			if err := store.SetValue(msg, COLUMN_SUBJECT, m.Subject); err != nil {
				log.Fatal("Unable set value:", err)
			}

			if err := store.SetValue(msg, COLUMN_DATE, m.Date.Format(dtFormat)); err != nil {
				log.Fatal("Unable set value:", err)
			}

			if err := store.SetValue(msg, COLUMN_FROM, strings.Join(m.From, "; ")); err != nil {
				log.Fatal("Unable set value:", err)
			}
			id := len(mails)
			m := m
			mails[id] = &m
			if err := store.SetValue(msg, COLUMN_OBJ, id); err != nil {
				log.Fatal("Unable set obj value:", err)
			}
		}
		/*
			if err := store.SetValue(msg, COLUMN_TO, m.To); err != nil {
				log.Fatal("Unable set value:", err)
			}
		*/
	}
}
