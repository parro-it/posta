// Package imap abstract away
// all imap code and provide access
// by mean of actions.
package imap

import (
	"context"
)

func Start(ctx context.Context) chan error {
	res := make(chan error)
	go func() {

		close(res)

	}()
	return res
}
