package msgbody

import (
	"context"
	"strings"

	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/chans"
	"github.com/parro-it/posta/imap"
	"github.com/parro-it/posta/msgs"
)

func Start(ctx context.Context) chan error {
	res := make(chan error)
	selectedMsg := app.ListenAction[msgs.MsgSelect]()

	go func() {
		defer close(res)

		for msgsel := range chans.WithContext(ctx, selectedMsg) {
			c, err := imap.AccountByName(msgsel.Msg.Account)
			if err != nil {
				panic(err)
			}
			err = c.FetchBody(msgsel.Msg)
			if err != nil {
				panic(err)
			}
			app.PostAction(MsgSetAll{
				Body:    msgsel.Msg.Body,
				Subject: msgsel.Msg.Subject,
				CC:      strings.Join(msgsel.Msg.CC, "; "),
				From:    strings.Join(msgsel.Msg.From, "; "),
				To:      strings.Join(msgsel.Msg.To, "; "),
				//Cc:      strings.Join(msgsel.Msg.Cc, "; "),
			})
		}
	}()
	return res
}

type MsgSetAll struct {
	Body    string
	Subject string
	From    string
	To      string
	CC      string
}
