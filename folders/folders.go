package folders

import (
	"context"
	"strings"
	"sync"

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
		var firstFolderLock sync.RWMutex
		var firstFolder string
		<-appStarted

		for _, account := range config.Values.Accounts {
			account := account

			g.Go(func() (err error) {
				qc := imapProc.QueryClient{
					Res:         make(chan *client.Client),
					AccountName: account.Name,
				}
				app.PostAction(qc)
				c := <-qc.Res

				ch := make(chan *imap.MailboxInfo)
				go func() {
					for f := range ch {
						path := strings.Split(f.Name, f.Delimiter)
						f := Folder{
							Name:    path[len(path)-1],
							Account: account.Name,
							Path:    path,
						}
						app.PostAction(Added{Folder: f})
						firstFolderLock.RLock()
						folder := firstFolder
						firstFolderLock.RUnlock()

						if folder == "" {
							firstFolderLock.Lock()
							firstFolder = f.Name
							firstFolderLock.Unlock()
							app.PostAction(Select{Folder: f})
						}
					}
				}()
				err = c.List("", "*", ch)
				return err
			})
		}
	}()
	return errs
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
