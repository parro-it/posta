package msgbody

import (
	"log"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/parro-it/posta/app"
)

// Creates a tree view and the tree store that holds its data
func View() *gtk.ScrolledWindow {
	text, err := gtk.TextViewNew()
	if err != nil {
		log.Fatal("Unable to create textview:", err)
	}
	scroll, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		log.Fatal("Unable to create scroolbox:", err)
	}
	scroll.Add(text)
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
