package plex_test

import (
	"sync"
	"testing"
	"time"

	"github.com/parro-it/posta/plex"
	"github.com/stretchr/testify/assert"
)

func TestDemux(t *testing.T) {
	t.Run("Send with 2 output channels", func(t *testing.T) {
		var q plex.Demux[int]

		q.Start()
		// close the demux once done
		defer q.Close()

		results1 := make(chan interface{})
		results2 := make(chan interface{})
		ints1 := plex.AddOut[int](q)
		ints2 := plex.AddOut[int](q)

		go func() {
			results1 <- <-ints1
			close(results1)
		}()
		go func() {
			results2 <- <-ints2
			close(results2)
		}()

		q.Input <- 42
		assert.Equal(t, 42, <-results1)
		assert.Equal(t, 42, <-results2)
	})

	t.Run("RemoveOut", func(t *testing.T) {
		var q plex.Demux[int]

		q.Start()
		// close the demux once done
		defer q.Close()

		results := make(chan interface{})
		ints := plex.AddOut[int](q)

		go func() {
			results <- <-ints
			close(results)
		}()

		// Pass non registered chans
		// to RemoveOut is a noop
		q.RemoveOut(make(chan struct{}))

		// this value will be read by the
		// goroutine above
		q.Input <- 42

		// this call should no block because
		// no output channel is present, after
		// ints is removed
		q.RemoveOut(ints)
		q.Input <- 43

		assert.Equal(t, 42, <-results)

	})

	t.Run("An unread output chan blocks all queue", func(t *testing.T) {
		var q plex.Demux[int]

		q.Start()

		// this output will
		// never be readed...
		plex.AddOut[int](q)

		// this send doesn't block,
		// but the start goroutine
		// will consequently get blocked
		// while forwarding the value to
		// an unread output channel
		q.Input <- 42

		select {
		case <-time.After(20 * time.Millisecond):
		case q.Input <- 42:
			assert.Fail(t, "The write is expected to timeout")
		}

		// this call would block because
		// the queue is already blocked.
		//q.Close()

	})

	t.Run("Send types with matching output doesn't block", func(t *testing.T) {
		var q plex.Demux[any]

		q.Start()
		await := make(chan struct{})
		go func() {
			actions := plex.AddOut[int](q)
			assert.Equal(t, 42, <-actions)
			close(await)
		}()
		<-time.After(20 * time.Millisecond)
		q.Input <- 42.42
		q.Input <- 42
		<-await
		q.Close()
	})

	t.Run("RemoveOut on two types", func(t *testing.T) {
		var q plex.Demux[any]

		q.Start()
		await := make(chan struct{})
		actions := plex.AddOut2[int, float64](q)

		go func() {
			assert.Equal(t, 42, <-actions)
			assert.Equal(t, 42.42, <-actions)
			close(await)
		}()
		// Unlisten non existent listeners
		// does nothing
		q.RemoveOut(make(chan struct{}))
		q.Input <- 42
		q.Input <- 42.42
		q.Input <- struct{}{}
		q.RemoveOut(actions)
		q.Input <- 42

		<-await
		q.Close()
	})

	t.Run("RemoveOut on 3 types", func(t *testing.T) {
		var q plex.Demux[any]

		q.Start()
		await := make(chan struct{})
		actions := plex.AddOut3[int, float64, bool](q)

		go func() {
			assert.Equal(t, 42, <-actions)
			assert.Equal(t, 42.42, <-actions)
			assert.Equal(t, true, <-actions)
			close(await)
		}()

		// Unlisten non existent listeners
		// does nothing
		q.RemoveOut(make(chan struct{}))

		q.Input <- 42
		q.Input <- 42.42
		q.Input <- true
		q.Input <- struct{}{}
		q.RemoveOut(actions)
		q.Input <- 42

		<-await
		q.Close()
	})

	t.Run("AddOut2", func(t *testing.T) {
		var q plex.Demux[any]

		q.Start()
		await := make(chan struct{})
		actions := plex.AddOut2[int, float64](q)

		go func() {
			assert.Equal(t, 42, <-actions)
			assert.Equal(t, 42.42, <-actions)
			close(await)
		}()
		q.Input <- 42
		q.Input <- 42.42
		q.Input <- struct{}{}
		<-await
		q.Close()
	})

	t.Run("AddOut3", func(t *testing.T) {
		var q plex.Demux[any]

		q.Start()
		await := make(chan struct{})
		actions := plex.AddOut3[int, float64, bool](q)

		go func() {
			assert.Equal(t, 42, <-actions)
			assert.Equal(t, 42.42, <-actions)
			assert.Equal(t, true, <-actions)
			close(await)
		}()
		q.Input <- 42
		q.Input <- 42.42
		q.Input <- true
		q.Input <- struct{}{}
		<-await
		q.Close()
	})

	t.Run("Each channel receive its own types", func(t *testing.T) {
		var q plex.Demux[any]

		q.Start()
		await := sync.WaitGroup{}
		await.Add(2)
		ints := plex.AddOut[int](q)
		floats := plex.AddOut[float64](q)

		go func() {
			assert.Equal(t, 42, <-ints)
			await.Done()
		}()
		go func() {
			assert.Equal(t, 42.42, <-floats)
			await.Done()
		}()
		<-time.After(20 * time.Millisecond)
		q.Input <- 42.42
		q.Input <- 42
		await.Wait()
		q.Close()
	})

}
