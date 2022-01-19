package msgs

import (
	"context"
	"log"

	imapProc "github.com/parro-it/posta/imap"

	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/chans"
	"github.com/parro-it/posta/folders"
)

type AddMsg struct {
	Msg imapProc.Msg
}

type ClearMsgs struct {
}

func Start(ctx context.Context) chan error {
	res := make(chan error)
	selectedFolders := app.ListenAction[folders.Select]()

	go func() {
		defer close(res)

		for fold := range chans.WithContext(ctx, selectedFolders) {
			app.PostAction(ClearMsgs{})
			c, err := imapProc.AccountByName(fold.Folder.Account)

			if err != nil {
				log.Printf("Cannot retrieve imap client: %s", err.Error())
				continue
			}

			msgs := c.ListMessages(imapProc.Folder{Path: fold.Folder.Path, Account: fold.Folder.Account})
			for msg := range msgs.Res {
				app.PostAction(AddMsg{Msg: msg})
			}

			if msgs.Err != nil {
				log.Printf("Cannot retrieve imap messages: %s", msgs.Err.Error())
				continue
			}

		}

		close(res)

	}()
	return res
}
