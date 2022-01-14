package msgs

import (
	"context"
	"log"
	"mime"
	"net/mail"
	"strings"

	imapProc "github.com/parro-it/posta/imap"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/folders"
)

type Msg struct {
	Date    string
	From    string
	To      string
	Subject string
}

type AddMsg struct {
	Msg Msg
}

func Start(ctx context.Context) chan error {
	res := make(chan error)
	selectedFolders := app.ListenAction[folders.Select]()

	go func() {
		defer close(res)

		for fold := range selectedFolders {

			qc := imapProc.QueryClient{
				Res:         make(chan *client.Client),
				AccountName: fold.Folder.Account,
			}
			app.PostAction(qc)
			c := <-qc.Res

			mbox, err := c.Select(strings.Join(fold.Folder.Path, "/"), false)
			if err != nil {
				log.Fatal(err)
			}

			// Get the last message
			if mbox.Messages == 0 {
				log.Fatal("No message in mailbox")
			}
			seqset := new(imap.SeqSet)
			seqset.AddRange(1, mbox.Messages)

			var section imap.BodySectionName
			section.Specifier = imap.HeaderSpecifier
			// Get the whole message body
			items := []imap.FetchItem{section.FetchItem()}

			messages := make(chan *imap.Message)
			done := make(chan error, 1)
			go func() {
				done <- c.Fetch(seqset, items, messages)
			}()
			log.Println("start")

			//log.Println("Last message:")
			for msg := range messages {
				r := msg.GetBody(&section)
				if r == nil {
					log.Fatal("Server didn't returned message body")
				}

				m, err := mail.ReadMessage(r)
				if err != nil {
					log.Fatal(err)
				}

				header := m.Header
				d, err := header.Date()
				if err != nil {
					panic(err)
				}
				from, err := header.AddressList("From")
				if err != nil {
					panic(err)
				}
				/*tos, err := header.AddressList("To")
				if err != nil {
					continue
				}
				to := strings.Builder{}
				for i, t := range tos {
					if i > 0 {
						to.WriteString(", ")
					}
					to.WriteString(t.Name)
				}
				*/

				dec := new(mime.WordDecoder)
				subject, err := dec.DecodeHeader(header.Get("Subject"))
				if err != nil {
					subject = err.Error() + " " + header.Get("Subject")
				}

				app.PostAction(AddMsg{Msg: Msg{
					Date: d.Format("2006/01/02 15:04"),
					From: from[0].Name,
					//To:      to.String(),
					Subject: subject,
				}})
				//log.Println("Date:", header.Get("Date"))
				//log.Println("From:", header.Get("From"))
				//log.Println("To:", header.Get("To"))
				//log.Println("Subject:", header.Get("Subject"))
			}
			if err := <-done; err != nil {
				log.Fatal(err)
			}
			log.Println("done")
			/*
				body, err := ioutil.ReadAll(m.Body)
				if err != nil {
					log.Fatal(err)
				}
				log.Println(body)
			*/
		}

		close(res)

	}()
	return res
}
