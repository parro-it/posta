package msgs

import (
	"context"
	"testing"

	"github.com/parro-it/posta/actions"
	"github.com/parro-it/posta/app"
	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	go actions.Start()

	errs := Start(context.Background())
	go app.Instance.Start()

	assert.NotEqual(t, 1, 2)

	err := <-errs
	assert.NoError(t, err)
}