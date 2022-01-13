package imap

import (
	"context"
	"testing"

	"github.com/emersion/go-imap/client"
	"github.com/parro-it/posta/app"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {

	app.Instance.Start(context.Background(), Client)

	q := QueryClient{
		Res:         make(chan *client.Client),
		AccountName: "cima",
	}
	app.PostAction(q)
	a := <-q.Res
	//assert.NoError(t, <-errs)
	assert.NotNil(t, a)
}
