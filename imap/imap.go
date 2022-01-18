// Package imap abstract away
// all imap code and provide access
// by mean of actions.
package imap

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"net/mail"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/chans"
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

// Account encapsulates a
// connection to an imap server
// and the configuration to make it.
type Account struct {
	Cfg    config.Account
	client *client.Client
}

type Folder struct {
	Size    uint32
	Name    string
	Account string
	Path    []string
	mbInfo  *imap.MailboxInfo
}

type Msg struct {
	Date    time.Time
	From    string
	To      []string
	Subject string
}

var accounts = map[string]Account{}

func AccountByName(name string) (*Account, error) {
	return nil, nil
}

type Result[T any] struct {
	Res chan T
	Err error
}

func (acc *Account) Login() *Result[struct{}] {
	res := Result[struct{}]{
		Res: make(chan struct{}),
	}
	go func() {
		var c *client.Client
		defer close(res.Res)

		if acc.Cfg.StartTLS {
			if c, res.Err = client.Dial(acc.Cfg.Addr); res.Err != nil {
				return
			}

			if res.Err = c.StartTLS(nil); res.Err != nil {
				return
			}

		} else {
			if c, res.Err = client.DialTLS(acc.Cfg.Addr, nil); res.Err != nil {
				return
			}
		}
		acc.client = c
		res.Err = c.Login(acc.Cfg.User, acc.Cfg.Pass)
	}()

	return &res
}

func (acc *Account) ListFolders() *Result[Folder] {
	res := Result[Folder]{
		Res: make(chan Folder),
	}
	ch := make(chan *imap.MailboxInfo)
	go func() {
		res.Err = acc.client.List("", "*", ch)
		//close(ch)
	}()
	go func() {
		defer close(res.Res)
		mboxes := chans.Collect(ch)
		for _, mb := range mboxes {
			var size uint32
			mbox, err := acc.client.Select(mb.Name, true)
			if err == nil {
				size = mbox.Messages
			}
			path := strings.Split(mb.Name, mb.Delimiter)
			res.Res <- Folder{
				Name:    path[len(path)-1],
				Account: acc.Cfg.Name,
				Path:    path,
				Size:    size,
				mbInfo:  mb,
			}
		}
	}()
	return &res
}

func (acc *Account) ListMessages(folder Folder) Result[Msg] {
	res := Result[Msg]{
		Res: make(chan Msg),
	}
	ch := make(chan *imap.Message)

	var section imap.BodySectionName
	section.Specifier = imap.HeaderSpecifier

	go func() {
		var mbox *imap.MailboxStatus

		if mbox, res.Err = acc.client.Select(strings.Join(folder.Path, "/"), true); res.Err != nil {
			close(ch)
			return
		}

		if mbox.Messages == 0 {
			close(ch)
			return
		}

		seqset := new(imap.SeqSet)
		seqset.AddRange(1, mbox.Messages)

		// Get the whole message body
		items := []imap.FetchItem{section.FetchItem()}
		res.Err = acc.client.Fetch(seqset, items, ch)
	}()
	go func() {
		defer close(res.Res)
		for msg := range ch {
			r := msg.GetBody(&section)
			if r == nil {
				res.Err = errors.New("Server didn't returned message body")
				return
			}
			out := Msg{}

			var m *mail.Message
			if m, res.Err = mail.ReadMessage(r); res.Err != nil {
				return
			}

			if out.Date, res.Err = m.Header.Date(); res.Err != nil {
				return
			}

			var from []*mail.Address

			if from, res.Err = m.Header.AddressList("From"); res.Err != nil {
				return
			}
			out.From = from[0].Name

			var to []*mail.Address

			if to, res.Err = m.Header.AddressList("To"); res.Err != nil {
				return
			}
			for _, a := range to {
				out.To = append(out.To, a.Name)
			}

			dec := new(mime.WordDecoder)
			subj := m.Header.Get("Subject")

			var e error
			if out.Subject, e = dec.DecodeHeader(subj); res.Err != nil {
				out.Subject = e.Error() + " " + subj
			}
			res.Res <- out
		}
	}()
	return res
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
