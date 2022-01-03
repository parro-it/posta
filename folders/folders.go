package folders

import (
	"context"
	"log"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/parro-it/posta/config"
	"golang.org/x/sync/errgroup"
)

type Folder struct {
	Name    string
	Account string
	Path    []string
}

type FoldersResults struct {
	Folders chan Folder
	Err     error
}

type Added struct {
	Folder Folder
}

type Removed struct {
	Folder Folder
}

func ReadFolders(ctx context.Context) *FoldersResults {
	g, ctx := errgroup.WithContext(ctx)

	var res FoldersResults
	res.Folders = make(chan Folder)

	for i, a := range config.Values.Accounts {
		i, a := i, a
		g.Go(func() (err error) {
			log.Printf("Connecting to server %s...", a.Addr)

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

			defer c.Logout()

			if err := c.Login(a.User, a.Pass); err != nil {
				return err
			}

			ch := make(chan *imap.MailboxInfo)
			go func() {
				for f := range ch {
					path := strings.Split(f.Name, f.Delimiter)
					f := Folder{
						Name:    path[len(path)-1],
						Account: a.Name,
						Path:    path,
					}
					res.Folders <- f
				}
			}()
			return c.List("", "*", ch)
		})
	}

	go func() {
		res.Err = g.Wait()
		close(res.Folders)
	}()

	return &res
}

/*
func main() {
	config.ParseCommandLine()
	err := config.Init()
	if err != nil {
		log.Fatal(err)
	}

	folders := ReadFolders(context.Background())
	for m := range folders.Folders {
		fmt.Println(m.Account, m.Path, m.Name)
	}

	if folders.Err != nil {
		log.Fatal(folders.Err)
	}

}
*/
