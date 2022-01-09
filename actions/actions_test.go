package actions_test

import (
	"testing"

	"github.com/parro-it/posta/actions"
)

func TestActions(t *testing.T) {
	go actions.Start()

}
