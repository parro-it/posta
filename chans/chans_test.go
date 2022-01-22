package chans_test

import (
	"context"
	"testing"

	"github.com/parro-it/posta/chans"
	"github.com/stretchr/testify/assert"
)

func TestWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in, _ := makeTestCh()
	ch := chans.WithContext(ctx, in)
	v, ok := <-ch
	assert.Equal(t, 0, v)
	assert.True(t, ok)

	v, ok = <-ch
	assert.Equal(t, 1, v)
	assert.True(t, ok)

	cancel()

	v, ok = <-ch
	assert.Equal(t, 0, v)
	assert.False(t, ok)
}

func TestCollect(t *testing.T) {
	t.Run("returns a slice", func(t *testing.T) {
		ch, expected := makeTestCh()

		actual := chans.Collect(ch)
		assert.Equal(t, expected, actual)
	})

	t.Run("with nil chan returns nil", func(t *testing.T) {
		actual := chans.Collect[int](nil)
		assert.Nil(t, actual)
	})

}

func makeTestCh() (chan int, []int) {
	ch := make(chan int, 10)
	var expected []int
	for i := 0; i < 10; i++ {
		ch <- i
		expected = append(expected, i)
	}
	close(ch)
	return ch, expected
}

func TestCollectIn(t *testing.T) {
	t.Run("returns a slice chan", func(t *testing.T) {
		ch, expected := makeTestCh()
		out := make(chan []int, 1)
		go chans.CollectIn(ch, out)
		assert.Equal(t, expected, <-out)
	})

	t.Run("with nil chan send nil", func(t *testing.T) {
		ch := make(chan []int, 1)
		chans.CollectIn(nil, ch)
		v, ok := <-ch
		assert.Nil(t, v)
		assert.True(t, ok)
		v, ok = <-ch
		assert.Nil(t, v)
		assert.False(t, ok)

	})

}
