package imap

import (
	"context"

	"github.com/emersion/go-imap/client"
	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/config"
	"github.com/parro-it/posta/plex"
)

type clientEntry struct {
	Account config.Account
	Client  *client.Client
}

// ClientManager is a component that
// manage connection and login to imap
// accounts. It also respond to a query
// action to get a connected client by
// name.
func Client(ctx context.Context) chan error {
	res := make(chan error)
	go func() {
		defer close(res)
		aa := plex.AddOut[QueryClient](app.Instance.Actions)
		<-plex.AddOut[app.AppStarted](app.Instance.Actions)

		clientsConfig := map[string]clientEntry{}

		// load all configured accounts from config
		// and map them by name
		for _, a := range config.Values.Accounts {
			clientsConfig[a.Name] = clientEntry{a, nil}
		}

		for a := range aa {
			ce, found := clientsConfig[a.AccountName]
			if !found {
				close(a.Res)
				continue
			}

			if ce.Client == nil {
				var err error
				if ce.Account.StartTLS {
					if ce.Client, err = client.Dial(ce.Account.Addr); err != nil {
						res <- err
						return
					}

					if err = ce.Client.StartTLS(nil); err != nil {
						res <- err
						return
					}

				} else {

					// Connect to server
					if ce.Client, err = client.DialTLS(ce.Account.Addr, nil); err != nil {
						res <- err
						return
					}
				}

			}

			a.Res <- ce.Client
			close(a.Res)

		}

	}()
	return res
}

type QueryClient struct {
	Res         chan *client.Client
	AccountName string
}
