package msgbody

import (
	"log"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/parro-it/posta/app"
)

// Creates a tree view and the tree store that holds its data
func View() *gtk.ScrolledWindow {

	ctrls, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	if err != nil {
		log.Fatal(err)
	}
	// from
	fromFlds, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	if err != nil {
		log.Fatal(err)
	}

	fromLbl, err := gtk.LabelNew("From:")
	if err != nil {
		log.Fatal(err)
	}
	fromLbl.SetWidthChars(8)
	fromLbl.SetXAlign(float64(gtk.ALIGN_START))

	fromFlds.PackStart(fromLbl, false, true, 0)

	from, err := gtk.EntryNew()
	if err != nil {
		log.Fatal(err)
	}
	fromFlds.PackStart(from, true, true, 2)
	ctrls.PackStart(fromFlds, false, false, 2)

	// to
	toFlds, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	if err != nil {
		log.Fatal(err)
	}

	toLbl, err := gtk.LabelNew("To:")
	if err != nil {
		log.Fatal(err)
	}
	toLbl.SetWidthChars(8)
	toLbl.SetXAlign(float64(gtk.ALIGN_START))

	toFlds.PackStart(toLbl, false, true, 0)

	to, err := gtk.EntryNew()
	if err != nil {
		log.Fatal(err)
	}
	toFlds.PackStart(to, true, true, 2)
	ctrls.PackStart(toFlds, false, false, 2)

	//cc
	ccFlds, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	if err != nil {
		log.Fatal(err)
	}
	ccLbl, err := gtk.LabelNew("CC:")
	if err != nil {
		log.Fatal(err)
	}
	ccLbl.SetWidthChars(8)
	ccLbl.SetXAlign(float64(gtk.ALIGN_START))

	ccFlds.PackStart(ccLbl, false, true, 0)

	cc, err := gtk.EntryNew()
	if err != nil {
		log.Fatal(err)
	}
	ccFlds.PackStart(cc, true, true, 2)
	ctrls.PackStart(ccFlds, false, false, 2)

	// subject
	subjFlds, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	if err != nil {
		log.Fatal(err)
	}
	subjLbl, err := gtk.LabelNew("Subject:")
	if err != nil {
		log.Fatal(err)
	}
	subjLbl.SetWidthChars(8)
	subjLbl.SetXAlign(float64(gtk.ALIGN_START))

	subjFlds.PackStart(subjLbl, false, true, 0)

	subj, err := gtk.EntryNew()
	if err != nil {
		log.Fatal(err)
	}
	subjFlds.PackStart(subj, true, true, 2)
	ctrls.PackStart(subjFlds, false, false, 2)

	//body
	text, err := gtk.TextViewNew()
	if err != nil {
		log.Fatal("Unable to create textview:", err)
	}
	ctrls.PackStart(text, true, true, 2)

	scroll, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		log.Fatal("Unable to create scroolbox:", err)
	}
	scroll.Add(ctrls)
	ch := app.ListenAction[MsgSetBody]()

	go func() {
		for setBody := range ch {
			setBody := setBody
			glib.IdleAdd(func() bool {
				b, err := text.GetBuffer()
				if err != nil {
					panic(err)
				}
				b.SetText(setBody.Text)
				return false
			})
		}
	}()
	return scroll
}
