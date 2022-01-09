package folders

import (
	"context"
	"strings"
	"sync"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/parro-it/posta/actions"
	"github.com/parro-it/posta/login"
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

const FOLDERS_SELECT = 4

func (s Select) Type() actions.ActionType {
	return FOLDERS_SELECT
}

type Added struct {
	Folder Folder
}

const FOLDERS_ADDED = 1

func (a Added) Type() actions.ActionType {
	return FOLDERS_ADDED
}

type Removed struct {
	Folder Folder
}

var clients []*client.Client

func Start(ctx context.Context) chan error {

	errs := make(chan error)
	go func() {
		g, _ := errgroup.WithContext(ctx)
		onClientReady := actions.ListenOf[login.ClientReady]()
		var firstFolderLock sync.RWMutex
		var firstFolder string

		for clientReady := range onClientReady {
			account := clientReady.Account
			c := clientReady.C

			g.Go(func() (err error) {

				ch := make(chan *imap.MailboxInfo)
				go func() {
					for f := range ch {
						path := strings.Split(f.Name, f.Delimiter)
						f := Folder{
							Name:    path[len(path)-1],
							Account: account,
							Path:    path,
						}
						actions.Post(Added{Folder: f})
						firstFolderLock.RLock()
						folder := firstFolder
						firstFolderLock.RUnlock()

						if folder == "" {
							firstFolderLock.Lock()
							firstFolder = f.Name
							firstFolderLock.Unlock()
							actions.Post(Select{Folder: f})
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
