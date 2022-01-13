package login

import (
	"context"
	"testing"

	"github.com/parro-it/posta/app"
	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {

	app.Instance.Start(context.Background(), Start)
	onClientReady := app.ListenAction[ClientReady]()

	c1 := <-onClientReady
	c2 := <-onClientReady

	assert.NotNil(t, c1)
	assert.NotNil(t, c2)
	assert.NotEqual(t, c1, c2)

	//err := <-errs
	//assert.NoError(t, err)
}
