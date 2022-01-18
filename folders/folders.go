package folders

import (
	"context"
	"fmt"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/config"
	imapProc "github.com/parro-it/posta/imap"
	"golang.org/x/sync/errgroup"
)

type Folder struct {
	Name    string
	Account string
	Path    []string
	//Imap    *imap.MailboxInfo
}

type Select struct {
	Folder Folder
}

type Added struct {
	Folder Folder
}

type Removed struct {
	Folder Folder
}

var clients []*client.Client

func Start(ctx context.Context) chan error {

	errs := make(chan error)
	appStarted := app.ListenAction[app.AppStarted]()

	go func() {
		g, _ := errgroup.WithContext(ctx)
		<-appStarted

		for i, account := range config.Values.Accounts {
			if i == 0 {
				g.Go(listClientFolder(account, true))

			} else {
				g.Go(listClientFolder(account, false))

			}

		}
	}()
	return errs
}

func listClientFolder(account config.Account, selFirstFolder bool) func() error {
	return func() (err error) {
		lf := imapProc.ListFolder{
			Res:         make(chan *imap.MailboxInfo),
			AccountName: account.Name,
		}

		qc := imapProc.QueryClient{
			Res:         make(chan *client.Client),
			AccountName: account.Name,
		}
		app.PostAction(qc)
		c := <-qc.Res

		app.PostAction(lf)
		for f := range lf.Res {
			var size uint32
			mbox, err := c.Select(f.Name, false)
			if err == nil {
				size = mbox.Messages
			}
			path := strings.Split(f.Name, f.Delimiter)
			f := Folder{
				Name:    fmt.Sprintf("%s (%d)", path[len(path)-1], size),
				Account: account.Name,
				Path:    path,
				//Imap:    f,
			}
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
