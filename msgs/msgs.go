package msgs

import (
	"context"
	"log"
	"net/mail"
	"strings"
	"sync"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/parro-it/posta/app"
	"github.com/parro-it/posta/folders"
	"github.com/parro-it/posta/login"
	"github.com/parro-it/posta/plex"
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
	var clientsSync sync.Mutex
	var clientsMap = map[string]*client.Client{}

	go func() {
		clients := plex.AddOut[login.ClientReady](app.Instance.Actions)
		for cr := range clients {
			clientsSync.Lock()
			clientsMap[cr.Account] = cr.C
			clientsSync.Unlock()
		}
	}()

	go func() {
		defer close(res)
		selectedFolders := plex.AddOut[folders.Select](app.Instance.Actions)

		for fold := range selectedFolders {
			clientsSync.Lock()
			c := clientsMap[fold.Folder.Account]
			clientsSync.Unlock()
			mbox, err := c.Select(strings.Join(fold.Folder.Path, "/"), false)
			if err != nil {
				log.Fatal(err)
			}

			// Get the last message
			if mbox.Messages == 0 {
				log.Fatal("No message in mailbox")
			}
			seqset := new(imap.SeqSet)
			seqset.AddRange(mbox.Messages-10, mbox.Messages)

			var section imap.BodySectionName
			section.Specifier = imap.HeaderSpecifier
			// Get the whole message body
			items := []imap.FetchItem{section.FetchItem()}

			messages := make(chan *imap.Message, 1)
			done := make(chan error, 1)
			go func() {
				done <- c.Fetch(seqset, items, messages)
			}()

			log.Println("Last message:")
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
				log.Println("Date:", header.Get("Date"))
				log.Println("From:", header.Get("From"))
				log.Println("To:", header.Get("To"))
				log.Println("Subject:", header.Get("Subject"))
				app.Instance.Actions.Input <- AddMsg{Msg: Msg{
					Date:    header.Get("Date"),
					From:    header.Get("From"),
					To:      header.Get("To"),
					Subject: header.Get("Subject"),
				}}
			}
			if err := <-done; err != nil {
				log.Fatal(err)
			}
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
