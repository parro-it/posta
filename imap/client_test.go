package imap

import (
	"context"
	"testing"

	"github.com/emersion/go-imap/client"
	"github.com/parro-it/posta/actions"
	"github.com/parro-it/posta/app"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	go actions.Start()
	errs := Client(context.Background())
	go app.Instance.Start()

	q := QueryClient{
		Res:         make(chan *client.Client),
		AccountName: "cima",
	}
	actions.Post(q)
	a := <-q.Res
	assert.NoError(t, <-errs)
	assert.NotNil(t, a)
}
