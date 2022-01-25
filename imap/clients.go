package imap

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/emersion/go-imap/client"
	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/config"
)

type BorrowedClient struct {
	*client.Client
	name string
}

func connectNewClient(account config.Account) (*client.Client, error) {
	var c *client.Client
	var err error

	if account.StartTLS {
		if c, err = client.Dial(account.Addr); err != nil {
			return nil, err
		}

		if err = c.StartTLS(nil); err != nil {
			c.Close()
			return nil, err
		}

	} else {
		if c, err = client.DialTLS(account.Addr, nil); err != nil {
			return nil, err
		}
	}

	err = c.Login(account.User, account.Pass)
	if err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

func AccountByName(name string) (*Account, error) {
	<-allClientsConnected
	a, ok := accounts[name]
	if !ok {
		return a, fmt.Errorf("Account not found with name %s", name)
	}
	return a, nil
}

func BorrowClient(accountName string) *BorrowedClient {
	<-allClientsConnected
	cmd := borrowClient{
		name: accountName,
		res:  make(chan *client.Client),
	}
	clientsCmds <- cmd
	return &BorrowedClient{<-cmd.res, accountName}
}
func (c *BorrowedClient) Done() {
	cmd := returnClient{
		name: c.name,
		c:    c.Client,
	}
	clientsCmds <- cmd
}
func distributeClients() {
	for _, a := range accounts {
		connect(a)
	}
	close(allClientsConnected)
	for c := range clientsCmds {
		switch cmd := c.(type) {
		case borrowClient:
			cc, ok := clients[cmd.name]
			if !ok || len(cc) == 0 {
				a := accounts[cmd.name]
				c, err := connectNewClient(a.Cfg)
				if err != nil {
					log.Fatalf("Cannot connect to %s: %s", a.Cfg.Name, err.Error())
				}
				cmd.res <- c
				close(cmd.res)
				continue
			}

			cmd.res <- cc[0]
			close(cmd.res)
			clients[cmd.name] = cc[1:]
		case returnClient:
			clients[cmd.name] = append(clients[cmd.name], cmd.c)
		}
	}
}

func connect(a *Account) {
	c, err := connectNewClient(a.Cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot connect to %s: %s", a.Cfg.Name, err.Error())
	}
	clients[a.Cfg.Name] = append(clients[a.Cfg.Name], c)
}

type borrowClient struct {
	name string
	res  chan *client.Client
}
type returnClient struct {
	name string
	c    *client.Client
}

var accounts = map[string]*Account{}
var clients = map[string][]*client.Client{}
var allClientsConnected = make(chan struct{})
var clientsCmds = make(chan any)

// ClientManager is a component that
// manage connection and login to imap
// accounts. It also respond to a query
// action to get a connected imapClient by
// name.
func Start(ctx context.Context) chan error {
	appStarted := app.ListenAction[app.AppStarted]()
	res := make(chan error)
	for _, a := range config.Values.Accounts {
		accounts[a.Name] = &Account{a}
	}
	go func() {
		defer close(res)
		<-appStarted
		// load all configured accounts from config
		// and map them by name
		distributeClients()
	}()
	return res
}
