package folders

import (
	"context"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/parro-it/posta/actions"
	"github.com/parro-it/posta/config"
	"golang.org/x/sync/errgroup"
)

type Folder struct {
	Name    string
	Account string
	Path    []string
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

func ReadFolders(ctx context.Context) chan error {
	errs := make(chan error)
	g, ctx := errgroup.WithContext(ctx)

	clientCh := make(chan *client.Client)
	for i, a := range config.Values.Accounts {
		i, a := i, a

		g.Go(func() (err error) {
			var c *client.Client
			if i == 0 {
				// Connect to server
				c, err = client.DialTLS(a.Addr, nil)
				if err != nil {
					return
				}
			} else {
				c, err = client.Dial(a.Addr)
				if err != nil {
					return
				}
				err = c.StartTLS(nil)
				if err != nil {
					return
				}
			}

			//defer c.Logout()

			if err := c.Login(a.User, a.Pass); err != nil {
				return err
			}

			clientCh <- c

			ch := make(chan *imap.MailboxInfo)
			go func() {
				for f := range ch {
					path := strings.Split(f.Name, f.Delimiter)
					f := Folder{
						Name:    path[len(path)-1],
						Account: a.Name,
						Path:    path,
					}
					actions.Post(Added{Folder: f})

				}
			}()
			err = c.List("", "*", ch)
			return err
		})
	}

	go func() {
		for i := 0; i < len(config.Values.Accounts); i++ {
			c := <-clientCh
			clients = append(clients, c)
		}
		close(clientCh)
		errs <- g.Wait()
		close(errs)
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
