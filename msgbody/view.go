package msgbody

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/errs"
)

// Creates a tree view and the tree store that holds its data
func View() *gtk.ScrolledWindow {

	ctrls := errs.Must(gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0))

	fromFlds := errs.Must(gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0))
	ctrls.PackStart(fromFlds, false, false, 2)

	fromLbl := errs.Must(gtk.LabelNew("From:"))
	fromLbl.SetWidthChars(8)
	fromLbl.SetXAlign(float64(gtk.ALIGN_START))
	fromFlds.PackStart(fromLbl, false, true, 0)

	from := errs.Must(gtk.EntryNew())
	fromFlds.PackStart(from, true, true, 2)

	// to
	toFlds := errs.Must(gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0))
	ctrls.PackStart(toFlds, false, false, 2)

	toLbl := errs.Must(gtk.LabelNew("To:"))
	toLbl.SetWidthChars(8)
	toLbl.SetXAlign(float64(gtk.ALIGN_START))
	toFlds.PackStart(toLbl, false, true, 0)

	to := errs.Must(gtk.EntryNew())
	toFlds.PackStart(to, true, true, 2)

	//cc
	ccFlds := errs.Must(gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0))
	ctrls.PackStart(ccFlds, false, false, 2)

	ccLbl := errs.Must(gtk.LabelNew("CC:"))
	ccLbl.SetWidthChars(8)
	ccLbl.SetXAlign(float64(gtk.ALIGN_START))
	ccFlds.PackStart(ccLbl, false, true, 0)

	cc := errs.Must(gtk.EntryNew())
	ccFlds.PackStart(cc, true, true, 2)

	// subject
	subjFlds := errs.Must(gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0))
	ctrls.PackStart(subjFlds, false, false, 2)

	subjLbl := errs.Must(gtk.LabelNew("Subject:"))
	subjLbl.SetWidthChars(8)
	subjLbl.SetXAlign(float64(gtk.ALIGN_START))
	subjFlds.PackStart(subjLbl, false, true, 0)

	subj := errs.Must(gtk.EntryNew())
	subjFlds.PackStart(subj, true, true, 2)

	//body
	text := errs.Must(gtk.TextViewNew())
	text.SetMarginBottom(10)
	text.SetMarginTop(10)
	text.SetMarginStart(10)
	text.SetMarginEnd(10)
	ctrls.PackStart(text, true, true, 2)

	scroll := errs.Must(gtk.ScrolledWindowNew(nil, nil))
	scroll.Add(ctrls)
	ch := app.ListenAction[MsgSetAll]()

	go func() {
		for setMsg := range ch {
			setMsg := setMsg
			glib.IdleAdd(func() bool {
				b, err := text.GetBuffer()
				if err != nil {
					panic(err)
				}
				b.SetText(setMsg.Body)
				subj.SetText(setMsg.Subject)
				from.SetText(setMsg.From)
				to.SetText(setMsg.To)
				cc.SetText(setMsg.CC)
				return false
			})
		}
	}()
	return scroll
}
