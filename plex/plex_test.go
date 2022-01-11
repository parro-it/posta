package plex_test

import (
	"sync"
	"testing"
	"time"

	"github.com/parro-it/posta/plex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDemux(t *testing.T) {

	t.Run("unlisten", func(t *testing.T) {
		q := make(plex.Demux)

		q.Start()
		await := make(chan struct{})
		actions := plex.AddOut[int](q)

		go func() {
			assert.Equal(t, 42, <-actions)
			close(await)
		}()

		// Unlisten non existent listeners
		// does nothing
		q.RemoveOut(make(chan struct{}))

		q <- 42
		q.RemoveOut(actions)
		// this call should no block because
		// no one is listening
		q <- 43

		<-await

		q.Close()
	})

	t.Run("use make as with channels", func(t *testing.T) {
		q := make(plex.Demux)

		require.NotNil(t, q)
	})

	t.Run("Start-Close", func(t *testing.T) {
		q := make(plex.Demux)

		q.Start()
		<-time.After(20 * time.Millisecond)
		q.Close()
	})

	t.Run("Send on single type", func(t *testing.T) {
		q := make(plex.Demux)

		q.Start()
		await := make(chan struct{})
		actions := plex.AddOut[int](q)

		go func() {
			assert.Equal(t, 42, <-actions)
			close(await)
		}()
		q <- 42
		<-await
		q.Close()
	})

	t.Run("declare the channel as var", func(t *testing.T) {
		var q plex.Demux

		q = q.Start()
		checkItSend(q, t, 42)
		q.Close()
	})

	t.Run("An unread listener block all queue", func(t *testing.T) {
		q := make(plex.Demux)

		q.Start()

		// this listener will
		// never be readed...
		plex.AddOut[int](q)

		// this send doesn't block,
		// but the start goroutine
		// will get blocked while sending
		// the value to an unreaded
		// listener channel
		q <- 42

		select {
		case <-time.After(20 * time.Millisecond):
		case q <- 42:
			assert.Fail(t, "The write is expected to timeout")
		}

		// this call would block because
		// the queue is already blocked.
		//q.Close()

	})

	t.Run("send types with no listener doesn't block", func(t *testing.T) {
		q := make(plex.Demux)

		q.Start()
		await := make(chan struct{})
		go func() {
			actions := plex.AddOut[int](q)
			assert.Equal(t, 42, <-actions)
			close(await)
		}()
		<-time.After(20 * time.Millisecond)
		q <- 42.42
		q <- 42
		<-await
		q.Close()
	})

	t.Run("unlisten on two types", func(t *testing.T) {
		q := make(plex.Demux)

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
		q <- 42
		q <- 42.42
		q <- struct{}{}
		q.RemoveOut(actions)
		q <- 42

		<-await
		q.Close()
	})

	t.Run("unlisten on 3 types", func(t *testing.T) {
		q := make(plex.Demux)

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

		q <- 42
		q <- 42.42
		q <- true
		q <- struct{}{}
		q.RemoveOut(actions)
		q <- 42

		<-await
		q.Close()
	})

	t.Run("listen two types", func(t *testing.T) {
		q := make(plex.Demux)

		q.Start()
		await := make(chan struct{})
		actions := plex.AddOut2[int, float64](q)

		go func() {
			assert.Equal(t, 42, <-actions)
			assert.Equal(t, 42.42, <-actions)
			close(await)
		}()
		q <- 42
		q <- 42.42
		q <- struct{}{}
		<-await
		q.Close()
	})

	t.Run("listen 3 types", func(t *testing.T) {
		q := make(plex.Demux)

		q.Start()
		await := make(chan struct{})
		actions := plex.AddOut3[int, float64, bool](q)

		go func() {
			assert.Equal(t, 42, <-actions)
			assert.Equal(t, 42.42, <-actions)
			assert.Equal(t, true, <-actions)
			close(await)
		}()
		q <- 42
		q <- 42.42
		q <- true
		q <- struct{}{}
		<-await
		q.Close()
	})

	t.Run("each listener receive its own types", func(t *testing.T) {
		q := make(plex.Demux)

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
		q <- 42.42
		q <- 42
		await.Wait()
		q.Close()
	})

}

func checkItSend[T any](q plex.Demux, t *testing.T, value T) {
	await := make(chan struct{})
	results := plex.AddOut[T](q)

	var res T
	go func() {
		res = <-results
		close(await)
	}()
	q <- value
	<-await
	assert.Equal(t, value, res)
}
