package login

import (
	"context"

	"github.com/emersion/go-imap/client"
	"github.com/parro-it/posta/actions"
	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/config"
	"golang.org/x/sync/errgroup"
)

type ClientReady struct {
	C       *client.Client
	Account string
}

const CLIENT_READY = 3

func (a ClientReady) Type() actions.ActionType {
	return CLIENT_READY
}

func Start(ctx context.Context) chan error {
	res := make(chan error)
	go func() {
		defer close(res)
		awaitStart := actions.Listen(app.APP_STARTED)
		<-awaitStart

		errs := make(chan error)
		g, _ := errgroup.WithContext(ctx)

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

				if err := c.Login(a.User, a.Pass); err != nil {
					return err
				}

				actions.Post(ClientReady{C: c, Account: a.Name})

				return err
			})
		}

		err := g.Wait()
		if err != nil {
			errs <- err
		}
		close(errs)

	}()
	return res
}
