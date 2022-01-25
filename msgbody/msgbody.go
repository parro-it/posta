package msgbody

import (
	"context"
	"strings"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/chans"
	"github.com/parro-it/posta/config"
	"github.com/parro-it/posta/imap"
	"github.com/parro-it/posta/msgs"
)

func Start(ctx context.Context) chan error {
	errs := make(chan error)
	actions := app.ListenAction2[msgs.MsgSelect, app.KeyPressed]()

	go func() {
		defer close(errs)
		var curmsg *imap.Msg

		setFields := func(editable bool) {
			app.PostAction(MsgSetAll{
				Attachments: curmsg.Attachments,
				Body:        curmsg.Body,
				Subject:     curmsg.Subject,
				From:        strings.Join(curmsg.From, "; "),
				To:          strings.Join(curmsg.To, "; "),
				CC:          strings.Join(curmsg.CC, "; "),
				Editable:    editable,
			})
		}
		app.Instance.RegisterShortcut(uint32(gdk.KEY_n), gdk.CONTROL_MASK, func() error {
			curmsg = &imap.Msg{
				Date:    time.Now(),
				From:    []string{config.Values.Accounts[0].User},
				To:      []string{},
				CC:      []string{},
				Subject: "",
				Body:    "",
				Account: config.Values.Accounts[0].Name,
				//Folder:      &imap.Folder{},
				Attachments: []imap.Attachment{},
			}
			setFields(true)
			return nil
		})
		for a := range chans.WithContext(ctx, actions) {
			switch action := a.(type) {
			case AttachmentsAdded:
				curmsg.Attachments = append(curmsg.Attachments, imap.Attachment{
					Name: action.Name,
				})
			case AttachmentsRemoved:
				for i, a := range curmsg.Attachments {
					if a.Name == action.Name {
						curmsg.Attachments = append(curmsg.Attachments[0:i], curmsg.Attachments[i+1:]...)
					}
				}
			case BodyChanged:
				curmsg.Body = action.Text
			case SubjectChanged:
				curmsg.Subject = action.Text
			case FromChanged:
				flds := strings.Split(action.Text, ";")
				for i := range flds {
					flds[i] = strings.TrimSpace(flds[i])
				}
				curmsg.From = flds
			case ToChanged:
				flds := strings.Split(action.Text, ";")
				for i := range flds {
					flds[i] = strings.TrimSpace(flds[i])
				}
				curmsg.To = flds
			case CCChanged:
				flds := strings.Split(action.Text, ";")
				for i := range flds {
					flds[i] = strings.TrimSpace(flds[i])
				}
				curmsg.CC = flds
			case app.KeyPressed:

			case msgs.MsgSelect:
				curmsg = action.Msg
				c, err := imap.AccountByName(curmsg.Account)
				if err != nil {
					panic(err)
				}
				err = c.FetchBody(curmsg)
				if err != nil {
					panic(err)
				}
				setFields(false)
			}
		}
	}()
	return errs
}

type MsgSetAll struct {
	Attachments []imap.Attachment
	Body        string
	Subject     string
	From        string
	To          string
	CC          string
	Editable    bool
}

type AttachmentsAdded struct {
	Name string
}
type AttachmentsRemoved struct {
	Name string
}
type BodyChanged struct {
	Text string
}
type SubjectChanged struct {
	Text string
}
type FromChanged struct {
	Text string
}
type ToChanged struct {
	Text string
}
type CCChanged struct {
	Text string
}
