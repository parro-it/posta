package msgbody

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/errs"
)

type flds struct {
	from *gtk.Entry
	to   *gtk.Entry
	cc   *gtk.Entry
	subj *gtk.Entry
	body *gtk.TextView
	atts *gtk.ListBox
}

// Creates a tree view and the tree store that holds its data
func View() *gtk.Frame {
	var flds flds

	var fromFlds *gtk.Box
	var toFlds *gtk.Box
	var ccFlds *gtk.Box
	var subjFlds *gtk.Box
	var bodyFlds *gtk.ScrolledWindow
	var left *gtk.Box
	var right *gtk.ScrolledWindow

	fromFlds, flds.from = fromCtrls()
	toFlds, flds.to = toCtrls()
	ccFlds, flds.cc = ccCtrls()
	subjFlds, flds.subj = subjCtrls()
	bodyFlds, flds.body = bodyCtrls()

	ctrls := errs.Must(gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0))

	left = errs.Must(gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0))
	left.PackStart(fromFlds, false, false, 2)
	left.PackStart(toFlds, false, false, 2)
	right, flds.atts = attachmentsCtrls()

	firstLines := errs.Must(gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0))
	firstLines.PackStart(left, true, true, 2)
	firstLines.PackStart(right, false, true, 2)

	ctrls.PackStart(firstLines, false, false, 2)
	ctrls.PackStart(ccFlds, false, false, 2)
	ctrls.PackStart(subjFlds, false, false, 2)
	ctrls.PackStart(bodyFlds, true, true, 2)

	scroll := errs.Must(gtk.FrameNew(""))
	scroll.Add(ctrls)

	go setFieldsOnAction(
		app.ListenAction[MsgSetAll](),
		flds,
	)

	return scroll
}

func setFieldsOnAction(ch chan MsgSetAll, flds flds) {
	for setMsg := range ch {
		setMsg := setMsg
		glib.IdleAdd(func() bool {
			b, err := flds.body.GetBuffer()
			if err != nil {
				panic(err)
			}
			b.SetText(setMsg.Body)
			flds.subj.SetText(setMsg.Subject)
			flds.from.SetText(setMsg.From)
			flds.to.SetText(setMsg.To)
			flds.cc.SetText(setMsg.CC)

			/*parent := errs.Must(flds.atts.GetParent()).(*gtk.Viewport)
			fld := errs.Must(gtk.ListBoxNew())
			fld.SetHAlign(gtk.ALIGN_FILL)
			fld.SetHExpand(true)
			fld.SetMarginBottom(3)
			fld.SetMarginStart(3)
			fld.SetMarginTop(3)
			fld.SetMarginEnd(3)

			flds.atts.Destroy()
			flds.atts = fld
			fld.ShowAll()
			parent.Add(fld)
			*/
			for {
				row := flds.atts.GetRowAtIndex(0)
				if row == nil {
					break
				}
				row.Destroy()
			}
			for _, a := range setMsg.Attachments {
				lbl := errs.Must(gtk.LabelNew(a.Name))
				lbl.SetHAlign(gtk.ALIGN_START)
				lbl.ShowAll()
				flds.atts.Add(lbl)
			}

			return false
		})
	}
}

func bodyCtrls() (*gtk.ScrolledWindow, *gtk.TextView) {
	text := errs.Must(gtk.TextViewNew())
	text.SetMarginBottom(10)
	text.SetMarginTop(10)
	text.SetMarginStart(10)
	text.SetMarginEnd(10)
	scroll := errs.Must(gtk.ScrolledWindowNew(nil, nil))
	scroll.Add(text)
	return scroll, text
}

func subjCtrls() (*gtk.Box, *gtk.Entry) {
	subjFlds := errs.Must(gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0))
	subjLbl := errs.Must(gtk.LabelNew("Subject:"))
	subjLbl.SetWidthChars(8)
	subjLbl.SetXAlign(float64(gtk.ALIGN_START))
	subjFlds.PackStart(subjLbl, false, true, 0)
	subj := errs.Must(gtk.EntryNew())
	subjFlds.PackStart(subj, true, true, 2)
	return subjFlds, subj
}

func attachmentsCtrls() (*gtk.ScrolledWindow, *gtk.ListBox) {
	ctrls := errs.Must(gtk.ScrolledWindowNew(nil, nil))
	fld := errs.Must(gtk.ListBoxNew())
	ctrls.Add(fld)
	ctrls.SetShadowType(gtk.SHADOW_OUT)
	ctrls.SetMinContentWidth(200)

	fld.SetHAlign(gtk.ALIGN_FILL)
	fld.SetHExpand(true)
	fld.SetMarginBottom(3)
	fld.SetMarginStart(3)
	fld.SetMarginTop(3)
	fld.SetMarginEnd(3)
	/*
		lbl := errs.Must(gtk.LabelNew("File.txt"))
		lbl.SetHAlign(gtk.ALIGN_START)
		fld.Add(lbl)
		lbl = errs.Must(gtk.LabelNew("File2.gif"))
		lbl.SetHAlign(gtk.ALIGN_START)
		fld.Add(lbl)
	*/
	return ctrls, fld
}

func ccCtrls() (*gtk.Box, *gtk.Entry) {
	ccFlds := errs.Must(gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0))
	ccLbl := errs.Must(gtk.LabelNew("CC:"))
	ccLbl.SetWidthChars(8)
	ccLbl.SetXAlign(float64(gtk.ALIGN_START))
	ccFlds.PackStart(ccLbl, false, true, 0)
	cc := errs.Must(gtk.EntryNew())
	ccFlds.PackStart(cc, true, true, 2)
	return ccFlds, cc
}

func toCtrls() (*gtk.Box, *gtk.Entry) {
	toFlds := errs.Must(gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0))
	toLbl := errs.Must(gtk.LabelNew("To:"))
	toLbl.SetWidthChars(8)
	toLbl.SetXAlign(float64(gtk.ALIGN_START))
	toFlds.PackStart(toLbl, false, true, 0)
	to := errs.Must(gtk.EntryNew())
	toFlds.PackStart(to, true, true, 2)
	return toFlds, to
}

func fromCtrls() (*gtk.Box, *gtk.Entry) {
	fromFlds := errs.Must(gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0))
	fromLbl := errs.Must(gtk.LabelNew("From:"))
	fromLbl.SetWidthChars(8)
	fromLbl.SetXAlign(float64(gtk.ALIGN_START))
	fromFlds.PackStart(fromLbl, false, true, 0)
	from := errs.Must(gtk.EntryNew())
	fromFlds.PackStart(from, true, true, 2)
	return fromFlds, from
}
