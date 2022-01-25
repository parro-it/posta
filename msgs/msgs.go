package msgs

import (
	"context"
	"log"

	"github.com/parro-it/posta/imap"

	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/chans"
	"github.com/parro-it/posta/folders"
)

type AddMsg struct {
	Msgs []imap.Msg
}

type ClearMsgs struct {
}

func Start(ctx context.Context) chan error {
	res := make(chan error)
	selectedFolders := app.ListenAction[folders.Select]()

	go func() {

		for fold := range chans.WithContext(ctx, selectedFolders) {
			app.PostAction(ClearMsgs{})
			c, err := imap.AccountByName(fold.Folder.Account)

			if err != nil {
				log.Printf("Cannot retrieve imap client: %s", err.Error())
				continue
			}

			msgs := c.ListMessages(ctx, imap.Folder{Path: fold.Folder.Path, Account: fold.Folder.Account})
			for msg := range chans.ChunksSplit(msgs.Res, 50) {
				app.PostAction(AddMsg{Msgs: msg})
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
