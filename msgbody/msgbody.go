package msgbody

import (
	"context"

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
			app.PostAction(MsgSetBody{
				Text: msgsel.Msg.Body,
			})
		}
	}()
	return res
}

type MsgSetBody struct {
	Text string
}
