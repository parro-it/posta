package chans_test

import (
	"io"
	"testing"

	"github.com/parro-it/posta/chans"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChanReader(t *testing.T) {
	type empty struct{}
	r, _ := chans.NewChanReader[empty]()
	var _ chans.Reader[empty] = r

	t.Run("Read", func(t *testing.T) {
		t.Run("2 values", func(t *testing.T) {
			ch, w := chans.NewChanReader(chans.PrefilledWith("uno", "due"))
			defer close(w)
			buf := make([]string, 2)
			n, err := ch.Read(buf)
			assert.NoError(t, err)
			assert.Equal(t, 2, n)
			assert.Equal(t, []string{"uno", "due"}, buf)
		})

		t.Run("ask more values then immediately present", func(t *testing.T) {
			ch, w := chans.NewChanReader(chans.PrefilledWith("uno", "due"))
			defer close(w)
			buf := make([]string, 3)
			n, err := ch.Read(buf)
			assert.NoError(t, err)
			assert.Equal(t, 2, n)
			assert.Equal(t, []string{"uno", "due", ""}, buf)
		})

		t.Run("ask more values then present, ch closed", func(t *testing.T) {
			ch, w := chans.NewChanReader(chans.PrefilledWith("uno", "due"))
			close(w)
			buf := make([]string, 3)
			n, err := ch.Read(buf)
			assert.EqualError(t, err, io.EOF.Error())
			assert.Equal(t, 2, n)
			assert.Equal(t, []string{"uno", "due", ""}, buf)
		})

		t.Run("ask less values then presents", func(t *testing.T) {
			ch, w := chans.NewChanReader(chans.PrefilledWith("uno", "due", "tre"))
			defer close(w)
			buf := make([]string, 2)
			n, err := ch.Read(buf)
			assert.NoError(t, err)
			assert.Equal(t, 2, n)
			assert.Equal(t, []string{"uno", "due"}, buf)
		})

		t.Run("ask less values then presents, then others", func(t *testing.T) {
			ch, w := chans.NewChanReader(chans.PrefilledWith("uno", "due", "tre"))
			buf := make([]string, 2)
			n, err := ch.Read(buf)
			assert.NoError(t, err)
			assert.Equal(t, 2, n)
			assert.Equal(t, []string{"uno", "due"}, buf)
			close(w)

			n, err = ch.Read(buf)
			assert.Equal(t, io.EOF, err)
			assert.Equal(t, 1, n)
			assert.Equal(t, []string{"tre", "due"}, buf)
		})

		t.Run("ask values on closed empty ch", func(t *testing.T) {
			ch, w := chans.NewChanReader[int]()
			buf := make([]int, 2)
			close(w)
			n, err := ch.Read(buf)
			assert.Equal(t, io.EOF, err)
			assert.Equal(t, 0, n)
			assert.Equal(t, []int{0, 0}, buf)
		})

		t.Run("ask values on empty ch", func(t *testing.T) {
			ch, w := chans.NewChanReader[int]()
			defer close(w)

			buf := make([]int, 2)
			n, err := ch.Read(buf)
			assert.NoError(t, err)
			assert.Equal(t, 0, n)
			assert.Equal(t, []int{0, 0}, buf)
		})

		t.Run("ask 0 values", func(t *testing.T) {
			ch, w := chans.NewChanReader[int]()
			defer close(w)
			buf := make([]int, 0)
			n, err := ch.Read(buf)
			assert.NoError(t, err)
			assert.Equal(t, 0, n)
			assert.Equal(t, []int{}, buf)
		})

		t.Run("ask 0 values on closed ch", func(t *testing.T) {
			ch, w := chans.NewChanReader[int]()
			close(w)
			buf := make([]int, 0)
			n, err := ch.Read(buf)
			assert.NoError(t, err)
			assert.Equal(t, 0, n)
			assert.Equal(t, []int{}, buf)
		})

		t.Run("pass nil buf", func(t *testing.T) {
			ch, w := chans.NewChanReader(chans.PrefilledWith(42))
			defer close(w)
			n, err := ch.Read(nil)
			assert.NoError(t, err)
			assert.Equal(t, 0, n)
		})

		t.Run("use nil chan", func(t *testing.T) {
			var ch chans.ChanReader[int]
			n, err := ch.Read(nil)
			buf := make([]int, 2)
			n, err = ch.Read(buf)
			assert.NoError(t, err)
			assert.Equal(t, 0, n)
		})
	})

	t.Run("creation with make", func(t *testing.T) {
		ch := make(chans.ChanReader[int])
		assert.NotNil(t, ch)

	})
	t.Run("NewChanReader", func(t *testing.T) {
		ch, w := chans.NewChanReader[int]()
		defer close(w)
		assert.NotNil(t, ch)
		assert.Equal(t, 0, cap(ch))

	})

	t.Run("NewPrefillChanReader", func(t *testing.T) {
		ch, w := chans.NewChanReader(chans.PrefilledWith(41, 42))
		defer close(w)
		assert.NotNil(t, ch)
		select {
		case v, ok := <-ch:
			require.True(t, ok)
			assert.Equal(t, 41, v)
		default:
			assert.Fail(t, "channel buffer is empty")
		}
		select {
		case v, ok := <-ch:
			require.True(t, ok)
			assert.Equal(t, 42, v)
		default:
			assert.Fail(t, "channel buffer is empty")
		}
	})
}
