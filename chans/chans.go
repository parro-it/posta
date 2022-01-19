package chans

import "context"

// Collect stores all values received from ch
// in a slice, and returns it when ch is closed.
func Collect[T any](ch chan T) []T {
	var res []T
	for v := range ch {
		res = append(res, v)
	}
	return res
}

// WithContext returns a channel that
// re-emits all values received from ch,
// but eventually get closed if the ctx context
// is cancelled.
func WithContext[T any](ctx context.Context, ch chan T) chan T {
	res := make(chan T)
	go func() {
		defer close(res)
		ctxCanceled := ctx.Done()
		for {
			select {
			case <-ctxCanceled:
				return
			case it, chOpen := <-ch:
				if !chOpen {
					return
				}
				res <- it
			}
		}
	}()
	return res
}
