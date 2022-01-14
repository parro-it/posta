// Package imap abstract away
// all imap code and provide access
// by mean of actions.
package imap

import (
	"context"
	"fmt"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/config"
)

type clientEntry struct {
	Account config.Account
	Client  *client.Client
}

type ClientReady struct {
	C       *client.Client
	Account string
}

// ClientManager is a component that
// manage connection and login to imap
// accounts. It also respond to a query
// action to get a connected client by
// name.
func Start(ctx context.Context) chan error {
	res := make(chan error)
	aa := app.ListenAction2[QueryClient, ListFolder]()
	appStarted := app.ListenAction[app.AppStarted]()

	go func() {
		defer close(res)
		<-appStarted

		clientsConfig := map[string]clientEntry{}

		// load all configured accounts from config
		// and map them by name
		for _, a := range config.Values.Accounts {
			clientsConfig[a.Name] = clientEntry{a, nil}
		}

		for a := range aa {
			switch action := a.(type) {
			case ListFolder:
				fmt.Printf("LISTFOLDER %s\n", action.AccountName)
				ce, found := clientsConfig[action.AccountName]
				if !found {
					panic(action.AccountName + " not found")
				}
				if ce.Client == nil {
					if err := connectClient(&ce); err != nil {
						res <- err
						return
					}
				}
				if err := ce.Client.List("", "*", action.Res); err != nil {
					res <- err
					return
				}
			case QueryClient:
				ce, found := clientsConfig[action.AccountName]
				if !found {
					close(action.Res)
					continue
				}

				if ce.Client == nil {
					if err := connectClient(&ce); err != nil {
						res <- err
						return
					}
				}

				action.Res <- ce.Client
				close(action.Res)
			}

		}

	}()
	return res
}

func connectClient(ce *clientEntry) error {
	var err error
	if ce.Account.StartTLS {
		if ce.Client, err = client.Dial(ce.Account.Addr); err != nil {
			return err
		}

		if err = ce.Client.StartTLS(nil); err != nil {
			return err
		}

	} else {

		if ce.Client, err = client.DialTLS(ce.Account.Addr, nil); err != nil {
			return err
		}
	}
	if err := ce.Client.Login(ce.Account.User, ce.Account.Pass); err != nil {
		return err
	}
	return nil
}

type QueryClient struct {
	Res         chan *client.Client
	AccountName string
}

type ListFolder struct {
	Res         chan *imap.MailboxInfo
	AccountName string
}
