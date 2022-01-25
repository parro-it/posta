package msgs

import (
	"context"
	"log"
	"sync"

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
		var cancelLock sync.Mutex

		runAction := func(ctx context.Context, fold folders.Select, cancel *context.CancelFunc) {
			app.PostAction(ClearMsgs{})
			c, err := imap.AccountByName(fold.Folder.Account)
			if err != nil {
				log.Printf("Cannot retrieve imap client: %s", err.Error())
				res <- err
				return
			}

			select {
			case <-ctx.Done():
				return
			default:
			}

			msgs := c.ListMessages(ctx, imap.Folder{Path: fold.Folder.Path, Account: fold.Folder.Account})
			for msg := range chans.WithContext(ctx, chans.ChunksSplit(msgs.Res, 50)) {
				app.PostAction(AddMsg{Msgs: msg})
			}

			if msgs.Err != nil {
				log.Printf("Cannot retrieve imap messages: %s", msgs.Err.Error())
				res <- msgs.Err
			}
			cancelLock.Lock()
			if *cancel != nil {
				(*cancel)()
				*cancel = nil
			}
			cancelLock.Unlock()

		}

		var cancel context.CancelFunc
		var actionsCtx context.Context

		for fold := range chans.WithContext(ctx, selectedFolders) {
			cancelLock.Lock()
			if cancel != nil {
				cancel()
				cancel = nil
			}
			actionsCtx, cancel = context.WithCancel(ctx)
			cancelLock.Unlock()

			go runAction(actionsCtx, fold, &cancel)

		}

		close(res)

	}()
	return res
}
