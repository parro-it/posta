package login

import (
	"context"
	"testing"

	"github.com/parro-it/posta/actions"
	"github.com/parro-it/posta/app"
	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	go actions.Start()
	onClientReady := actions.Listen(CLIENT_READY)

	errs := Start(context.Background())
	go app.Instance.Start()

	c1 := <-onClientReady
	c2 := <-onClientReady

	assert.NotNil(t, c1)
	assert.NotNil(t, c2)
	assert.NotEqual(t, c1, c2)

	err := <-errs
	assert.NoError(t, err)
}