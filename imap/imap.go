// Package imap abstract away
// all imap code and provide access
// by mean of actions.
package imap

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/mail"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-message"

	"github.com/parro-it/posta/chans"
	"github.com/parro-it/posta/config"
	"github.com/parro-it/posta/errs"

	md "github.com/JohannesKaufmann/html-to-markdown"
)

// Account encapsulates a
// connection to an imap server
// and the configuration to make it.
type Account struct {
	Cfg config.Account
	//imapClient *client.Client
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

func (acc *Account) FetchBody(ctx context.Context, msg *Msg) error {
	c := BorrowClient(acc.Cfg.Name)
	defer c.Done()

	var err error
	if _, err = c.Select(msg.Folder.Path, true); err != nil {
		return err
	}

	seqset := new(imap.SeqSet)
	seqset.AddRange(msg.Uid, msg.Uid)

	// Get the whole message body
	var section imap.BodySectionName

	items := []imap.FetchItem{section.FetchItem()}
	var ch chan *imap.Message
	//for {
	ch = make(chan *imap.Message, 1)
	/*
		err = c.Fetch(seqset, items, ch)
			if err == nil {
				break
			}
			if err.Error() == "short write" {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			return err
		}*/
	if err = c.Fetch(seqset, items, ch); err != nil {
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
		log.Println("Unknown charset:", err)
	} else if err != nil {
		return err
	}
	bs, err := fetchBodyStructure(m)
	if err != nil {
		return err
	}

	msg.Attachments = bs.Attachments
	if bs.textContent != nil {
		msg.Body = readText(bs.textContent, bs.charset)
	} else if bs.htmlContent != nil {
		msg.Body, err = readHMTL(bs.htmlContent)
		return err
	}

	return nil
}

func (acc *Account) ListFolders(ctx context.Context) errs.Result[Folder] {
	res := errs.Result[Folder]{
		Res: make(chan Folder),
	}
	c := BorrowClient(acc.Cfg.Name)

	ch := make(chan *imap.MailboxInfo)
	go func() {
		res.Err = c.List("", "*", ch)
		c.Done()
		//close(ch)
	}()
	go func() {
		defer close(res.Res)
		mboxes := chans.Collect(ch)
		for _, mb := range mboxes {
			var size uint32
			mbox, err := c.Select(mb.Name, true)
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
	return res
}

func (acc *Account) ListMessages(ctx context.Context, folder Folder) errs.Result[Msg] {
	res := errs.NewResult[Msg]()

	var chunkSz uint32 = 50
	var mbox *imap.MailboxStatus
	c := BorrowClient(acc.Cfg.Name)

	// this function fetch a chunk of messages using
	// Fetch imap command, sending them to an out chan
	// chunkSeq is the sequence number of the chunk, starting with 0
	// out is the output channel.
	// The function closes out chan when it has done sending all msgs
	fetchChunk := func(chunkSeq uint32, out chan *imap.Message) {
		seqset := new(imap.SeqSet)
		start := chunkSeq*chunkSz + 1
		stop := (chunkSeq + 1) * chunkSz
		if stop > mbox.Messages {
			stop = mbox.Messages
		}
		seqset.AddRange(start, stop)

		// Get the whole message body
		items := []imap.FetchItem{imap.FetchEnvelope}
		res.Err = c.Fetch(seqset, items, out)
	}

	go func() {
		defer c.Done()
		defer close(res.Res)

		if mbox, res.Err = c.Select(folder.Path, true); res.Err != nil {
			return
		}

		if mbox.Messages == 0 {
			return
		}
		//var mux = chans.SimpleMux[*imap.Message]{Output: msgChan}
		//defer mux.Close()
		for i := uint32(0); i <= mbox.Messages/chunkSz; i++ {
			ch := make(chan *imap.Message, chunkSz)
			//mux.AddInputFrom(ch)

			fetchChunk(i, ch)
			if res.Err != nil {
				return
			}

			for msg := range ch {
				res.Res <- msgFromImap(msg, acc, folder)
				select {
				case <-ctx.Done():
					fmt.Printf("ListMessages on folder `%s` canceled.\n", folder.Path)
					return
				default:
					continue
				}
			}

			select {
			case <-ctx.Done():
				fmt.Printf("ListMessages on folder `%s` canceled.\n", folder.Path)
				return
			default:
				continue
			}
		}

	}()

	return res
}

func msgFromImap(msg *imap.Message, acc *Account, folder Folder) Msg {
	en := msg.Envelope

	out := Msg{
		Uid:     msg.SeqNum,
		Account: acc.Cfg.Name,
		Folder:  &folder,
		Date:    en.Date,
		Subject: en.Subject,
	}
	out.From = formatAddresses(en.From)
	out.To = formatAddresses(en.To)
	out.CC = formatAddresses(en.Cc)

	return out
}

type bodyStructure struct {
	textContent []byte
	htmlContent []byte
	charset     string
	Attachments []Attachment
}

func fetchBodyStructure(m *message.Entity) (bodyStructure, error) {
	var bs bodyStructure
	mr := m.MultipartReader()
	if mr == nil {
		// non MIME mail
		t, params, err := m.Header.ContentType()
		if err != nil {
			return bs, err
		}
		bs.charset = params["charset"]
		if t == "text/plain" {
			bs.textContent, err = io.ReadAll(m.Body)
			if err != nil {
				return bs, err
			}
		} else if t == "text/html" {
			bs.htmlContent, err = io.ReadAll(m.Body)
			if err != nil {
				return bs, err
			}
		}
		return bs, nil
	}

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			return bs, err
		}

		t, typeParams, err := p.Header.ContentType()
		if err != nil {
			return bs, err
		}
		disp, params, err := p.Header.ContentDisposition()
		if err != nil && err.Error() != "mime: no media type" {
			return bs, err
		}

		if disp == "attachment" {
			att := Attachment{
				Name:   params["filename"],
				reader: p.Body,
			}
			bs.Attachments = append(bs.Attachments, att)
			continue
		}
		if t == "text/plain" {
			bs.charset = typeParams["charset"]
			bs.textContent = errs.Must(io.ReadAll(p.Body))
		} else if t == "text/html" {
			bs.charset = typeParams["charset"]
			bs.htmlContent = errs.Must(io.ReadAll(p.Body))
		}

	}
	return bs, nil
}

func readText(r []byte, charset string) string {
	return string(r)
}
func readHMTL(r []byte) (string, error) {

	converter := md.NewConverter("", true, nil)

	markdown, err := converter.ConvertBytes(r)
	if err != nil {
		return "", err
	}
	return string(markdown), nil
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
