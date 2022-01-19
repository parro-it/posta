// Package imap abstract away
// all imap code and provide access
// by mean of actions.
package imap

import (
	"context"
	"fmt"
	"io"
	"log"
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

// Account encapsulates a
// connection to an imap server
// and the configuration to make it.
type Account struct {
	Cfg    config.Account
	client *client.Client
}

type Folder struct {
	Sep     string
	Size    uint32
	Name    string
	Account string
	Path    string
	mbInfo  *imap.MailboxInfo
}

type Msg struct {
	Uid     uint32
	Date    time.Time
	From    string
	To      []string
	Subject string
	Body    string
	mail    *mail.Message
	Account string
	Folder  *Folder
}

var accounts = map[string]*Account{}

func AccountByName(name string) (*Account, error) {
	a, ok := accounts[name]
	if !ok {
		return a, fmt.Errorf("Account not found with name %s", name)
	}
	return a, nil
}

type Result[T any] struct {
	Res chan T
	Err error
}

func (acc *Account) FetchBody(msg *Msg) error {

	var section imap.BodySectionName
	section.Specifier = imap.EntireSpecifier

	var err error
	if _, err = acc.client.Select(msg.Folder.Path, true); err != nil {
		return err
	}

	seqset := new(imap.SeqSet)
	seqset.AddRange(msg.Uid, msg.Uid)

	// Get the whole message body
	items := []imap.FetchItem{section.FetchItem()}
	ch := make(chan *imap.Message, 1)
	if err = acc.client.Fetch(seqset, items, ch); err != nil {
		return err
	}

	res := <-ch
	r := res.GetBody(&section)
	if r == nil {
		return fmt.Errorf("Server didn't returned message body")
	}
	/*
		var m *mail.Message
		if m, err = mail.ReadMessage(r); err != nil {
			return fmt.Errorf("Cannot read message: %s\n", err.Error())
		}
	*/
	body, err := io.ReadAll(r /*m.Body*/)
	if err != nil {
		return err
	}
	msg.Body = string(body)
	return nil
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
				Path:    mb.Name,
				Sep:     mb.Delimiter,
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

		if mbox, res.Err = acc.client.Select(folder.Path, true); res.Err != nil {
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
				log.Println("Server didn't returned message body")
				continue
			}
			out := Msg{
				Uid:     msg.SeqNum,
				Account: acc.Cfg.Name,
				Folder:  &folder,
			}

			var m *mail.Message
			var err error
			if m, err = mail.ReadMessage(r); err != nil {
				log.Printf("Cannot read message: %s\n", err.Error())
				continue
			}
			out.mail = m
			if out.Date, err = m.Header.Date(); err != nil {
				log.Printf("Cannot read message date: %s\n", err.Error())
				continue
			}

			var from []*mail.Address

			if from, err = m.Header.AddressList("From"); err != nil {
				log.Printf("Cannot read message From: %s\n", err.Error())
				continue
			}
			out.From = from[0].Name

			var to []*mail.Address

			if to, err = m.Header.AddressList("To"); err != nil {
				log.Printf("Cannot read message From: %s\n", err.Error())
				continue
			}
			for _, a := range to {
				out.To = append(out.To, a.Name)
			}

			dec := new(mime.WordDecoder)
			subj := m.Header.Get("Subject")

			if out.Subject, err = dec.DecodeHeader(subj); err != nil {
				log.Printf("Cannot decode subject: %s\n", err.Error())
				out.Subject = subj
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
	appStarted := app.ListenAction[app.AppStarted]()
	res := make(chan error)
	go func() {
		defer close(res)
		<-appStarted

		// load all configured accounts from config
		// and map them by name
		for _, a := range config.Values.Accounts {
			accounts[a.Name] = &Account{a, nil}
		}
	}()
	return res
}
