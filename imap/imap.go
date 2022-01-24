// Package imap abstract away
// all imap code and provide access
// by mean of actions.
package imap

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/mail"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message"

	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/chans"
	"github.com/parro-it/posta/config"

	md "github.com/JohannesKaufmann/html-to-markdown"
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
	Uid         uint32
	Date        time.Time
	From        []string
	To          []string
	CC          []string
	Subject     string
	Body        string
	mail        *mail.Message
	Account     string
	Folder      *Folder
	Attachments []Attachment
}

type Attachment struct {
	reader io.Reader
	Name   string
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

	var err error
	if _, err = acc.client.Select(msg.Folder.Path, true); err != nil {
		return err
	}

	seqset := new(imap.SeqSet)
	seqset.AddRange(msg.Uid, msg.Uid)

	// Get the whole message body
	var section imap.BodySectionName

	items := []imap.FetchItem{section.FetchItem()}
	ch := make(chan *imap.Message, 1)
	if err = acc.client.Fetch(seqset, items, ch); err != nil {
		return err
	}

	res := <-ch
	/*s, err := json.MarshalIndent(res.BodyStructure, "  ", "  ")
	if err != nil {
		panic(err)
	}*/
	r := res.GetBody(&section)
	m, err := message.Read(r)
	if message.IsUnknownCharset(err) {
		// This error is not fatal
		log.Println("Unknown encoding:", err)
	} else if err != nil {
		log.Fatal(err)
	}
	var s string
	msg.Attachments = []Attachment{}
	if mr := m.MultipartReader(); mr != nil {
		// This is a multipart message
		s = "This is a multipart message containing:\n"
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}

			disp, params, err := p.Header.ContentDisposition()
			if disp == "attachment" {
				att := Attachment{
					Name:   params["filename"],
					reader: p.Body,
				}
				msg.Attachments = append(msg.Attachments, att)
				continue
			}

			t, _, _ := p.Header.ContentType()
			if t == "text/plain" {
				s = readText(p.Body)
				break
			} else if t == "text/html" {
				s = readHMTL(p.Body)
				break
			}
			s += fmt.Sprintf("\t- A part with type %s \n", t)
		}
	} else {
		t, _, _ := m.Header.ContentType()
		if t == "text/plain" {
			s = readText(m.Body)
		} else if t == "text/html" {
			s = readHMTL(m.Body)
		} else {
			s = fmt.Sprintf("This is a non-multipart message with type %s \n", t)
		}
	}

	msg.Body = s
	return nil
}

func readText(r io.Reader) string {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		log.Fatal(err)
	}
	return buf.String()
}
func readHMTL(r io.Reader) string {

	converter := md.NewConverter("", true, nil)

	markdown, err := converter.ConvertReader(r)
	if err != nil {
		log.Fatal(err)
	}
	return markdown.String()
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
		items := []imap.FetchItem{imap.FetchEnvelope}
		res.Err = acc.client.Fetch(seqset, items, ch)
	}()

	go func() {
		defer close(res.Res)
		//dec := new(mime.WordDecoder)

		for msg := range ch {

			/*r := msg.GetBody(&section)
			if r == nil {
				log.Println("Server didn't returned message body")
				continue
			}*/
			en := msg.Envelope

			out := Msg{
				Uid:     msg.SeqNum,
				Account: acc.Cfg.Name,
				Folder:  &folder,
				Date:    en.Date,
				Subject: en.Subject,
			}

			/*fd, err := dec.Decode(out.From)
			if err == nil {
				out.From = fd
			}*/
			out.From = formatAddresses(en.From)
			out.To = formatAddresses(en.To)
			out.CC = formatAddresses(en.Cc)

			res.Res <- out
		}
	}()
	return res
}

func formatAddresses(addrs []*imap.Address) []string {
	var res []string
	for _, a := range addrs {
		var s string
		if a.PersonalName != "" {
			s = a.PersonalName
		} else {
			s = a.Address()
		}

		res = append(res, s)
	}
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
