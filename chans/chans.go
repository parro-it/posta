package chans

import (
	"context"
)

// CollectIn stores all values received from ch
// in a slice, and send the slice to out channel
// when ch is closed. The out channel is closed afterwards.
func CollectIn[T any](ch chan T, out chan []T) {
	out <- Collect(ch)
	close(out)
}

// Collect stores all values received from ch
// in a slice, and returns it when ch is closed.
func Collect[T any](ch chan T) []T {
	if ch == nil {
		return nil
	}
	var res []T
	for v := range ch {
		res = append(res, v)
	}
	return res
}

// ChunksSplit groups values received from ch
// in chunks of chunkLen size, and send each chunk to a
// channel of slices.
func ChunksSplit[T any](ch chan T, chunkSz int) chan []T {
	if ch == nil {
		return nil
	}
	res := make(chan []T)
	go func() {
		for {
			chunk := []T{}
			var v T
			isOpen := true
			for i := 0; isOpen && i < chunkSz; i++ {
				v, isOpen = <-ch
				if !isOpen {
					break
				}

				chunk = append(chunk, v)
			}
			res <- chunk
			if !isOpen {
				break
			}

		}
	}()
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
