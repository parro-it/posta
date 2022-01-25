package folders

import (
	"context"

	"github.com/emersion/go-imap/client"
	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/chans"
	"github.com/parro-it/posta/config"
	"github.com/parro-it/posta/imap"
	"golang.org/x/sync/errgroup"
)

/*
type Folder struct {
	Name    string
	Account string
	Path    []string
	//Imap    *imap.MailboxInfo
}
*/
type Select struct {
	Folder imap.Folder
}

type Added struct {
	Folder imap.Folder
}

type Removed struct {
	Folder imap.Folder
}

var clients []*client.Client

func Start(ctx context.Context) chan error {

	errs := make(chan error)
	appStarted := app.ListenAction[app.AppStarted]()

	go func() {
		g, _ := errgroup.WithContext(ctx)
		<-appStarted

		for i, account := range config.Values.Accounts {
			if i == 1 {
				g.Go(listClientFolder(ctx, account, true))
			} else {
				g.Go(listClientFolder(ctx, account, false))
			}
		}
		errs <- g.Wait()
	}()
	return errs
}

func listClientFolder(ctx context.Context, account config.Account, selFirstFolder bool) func() error {
	return func() (err error) {

		c, err := imap.AccountByName(account.Name)
		if err != nil {
			return err
		}
		//res := c.Login()
		//<-res.Res
		//if res.Err != nil {
		//	panic(res.Err)
		//}

		lf := c.ListFolders(ctx)
		for f := range chans.WithContext(ctx, lf.Res) {

			app.PostAction(Added{Folder: f})

			if selFirstFolder && f.Name == "INBOX" {
				app.PostAction(Select{Folder: f})
				selFirstFolder = false
			}
		}
		return err
	}
}

/*
func ListenUpdates(ctx context.Context) chan error {
	errs := make(chan error)

	for _, c := range clients {
		c := c
		updates := make(chan client.Update)
		c.Updates = updates
		stop := make(chan struct{})
		done := make(chan error, 1)
		var stopped bool
		go func() {
			done <- c.Idle(stop, nil)
		}()
		for {
			select {
			case update := <-updates:
				log.Println("New update:", update)
				if !stopped {
					close(stop)
					stopped = true
				}
			case err := <-done:
				if err != nil {
					log.Fatal(err)
				}
				log.Println("Not idling anymore")
				return errs
			}
		}
	}
	return errs
}
*/
