package msgbody

import (
	"context"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/chans"
	"github.com/parro-it/posta/imap"
	"github.com/parro-it/posta/msgs"
)

func Start(ctx context.Context) chan error {
	res := make(chan error)
	actions := app.ListenAction2[msgs.MsgSelect, app.KeyPressed]()

	go func() {
		defer close(res)

		for a := range chans.WithContext(ctx, actions) {
			switch action := a.(type) {
			case app.KeyPressed:
				if action.Key == gdk.KEY_n && action.State&gdk.CONTROL_MASK == gdk.CONTROL_MASK {
					app.PostAction(MsgSetAll{Editable: true})
				}
			case msgs.MsgSelect:
				c, err := imap.AccountByName(action.Msg.Account)
				if err != nil {
					panic(err)
				}
				err = c.FetchBody(action.Msg)
				if err != nil {
					panic(err)
				}
				app.PostAction(MsgSetAll{
					Attachments: action.Msg.Attachments,
					Body:        action.Msg.Body,
					Subject:     action.Msg.Subject,
					CC:          strings.Join(action.Msg.CC, "; "),
					From:        strings.Join(action.Msg.From, "; "),
					To:          strings.Join(action.Msg.To, "; "),
				})
			}
		}
	}()
	return res
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
