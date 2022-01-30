package chans_test

import (
	"io"
	"testing"

	"github.com/parro-it/posta/chans"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThrough(t *testing.T) {
	type empty struct{}
	var _ io.Closer = chans.NewThrough[empty]()
	var _ chans.Reader[empty] = chans.NewThrough[empty]()
	var _ chans.Writer[empty] = chans.NewThrough[empty]()
	t.Run("Write", func(t *testing.T) {
		t.Run("2 values", func(t *testing.T) {
			ch := chans.NewBufThrough[int](2)
			n, err := ch.Write([]int{41, 42})
			assert.Equal(t, 2, n)
			assert.NoError(t, err)
			assert.Equal(t, 41, <-ch)
			assert.Equal(t, 42, <-ch)
			ch.Close()
			_, ok := <-ch
			assert.False(t, ok)
		})
	})
	t.Run("Read", func(t *testing.T) {
		t.Run("2 values", func(t *testing.T) {
			ch := chans.NewPrefillThrough("uno", "due")
			defer ch.Close()
			buf := make([]string, 2)
			n, err := ch.Read(buf)
			assert.NoError(t, err)
			assert.Equal(t, 2, n)
			assert.Equal(t, []string{"uno", "due"}, buf)
		})

		t.Run("ask more values then immediately present", func(t *testing.T) {
			ch := chans.NewPrefillThrough("uno", "due")
			defer ch.Close()
			buf := make([]string, 3)
			n, err := ch.Read(buf)
			assert.NoError(t, err)
			assert.Equal(t, 2, n)
			assert.Equal(t, []string{"uno", "due", ""}, buf)
		})

		t.Run("ask more values then present, ch closed", func(t *testing.T) {
			ch := chans.NewPrefillThrough("uno", "due")
			ch.Close()
			buf := make([]string, 3)
			n, err := ch.Read(buf)
			assert.EqualError(t, err, io.EOF.Error())
			assert.Equal(t, 2, n)
			assert.Equal(t, []string{"uno", "due", ""}, buf)
		})

		t.Run("ask less values then presents", func(t *testing.T) {
			ch := chans.NewPrefillThrough("uno", "due", "tre")
			defer ch.Close()
			buf := make([]string, 2)
			n, err := ch.Read(buf)
			assert.NoError(t, err)
			assert.Equal(t, 2, n)
			assert.Equal(t, []string{"uno", "due"}, buf)
		})

		t.Run("ask less values then presents, then others", func(t *testing.T) {
			ch := chans.NewPrefillThrough("uno", "due", "tre")
			buf := make([]string, 2)
			n, err := ch.Read(buf)
			assert.NoError(t, err)
			assert.Equal(t, 2, n)
			assert.Equal(t, []string{"uno", "due"}, buf)
			ch.Close()

			n, err = ch.Read(buf)
			assert.Equal(t, io.EOF, err)
			assert.Equal(t, 1, n)
			assert.Equal(t, []string{"tre", "due"}, buf)
		})

		t.Run("ask values on closed empty ch", func(t *testing.T) {
			ch := chans.NewThrough[int]()
			ch.Close()
			buf := make([]int, 2)

			n, err := ch.Read(buf)
			assert.Equal(t, io.EOF, err)
			assert.Equal(t, 0, n)
			assert.Equal(t, []int{0, 0}, buf)
		})

		t.Run("ask values on empty ch", func(t *testing.T) {
			ch := chans.NewThrough[int]()
			defer ch.Close()
			buf := make([]int, 2)
			n, err := ch.Read(buf)
			assert.NoError(t, err)
			assert.Equal(t, 0, n)
			assert.Equal(t, []int{0, 0}, buf)
		})

		t.Run("ask 0 values", func(t *testing.T) {
			ch := chans.NewThrough[int]()
			defer ch.Close()
			buf := make([]int, 0)
			n, err := ch.Read(buf)
			assert.NoError(t, err)
			assert.Equal(t, 0, n)
			assert.Equal(t, []int{}, buf)
		})

		t.Run("ask 0 values on closed ch", func(t *testing.T) {
			ch := chans.NewThrough[int]()
			ch.Close()
			buf := make([]int, 0)
			n, err := ch.Read(buf)
			assert.NoError(t, err)
			assert.Equal(t, 0, n)
			assert.Equal(t, []int{}, buf)
		})

		t.Run("pass nil buf", func(t *testing.T) {
			ch := chans.NewPrefillThrough(42)
			defer ch.Close()
			n, err := ch.Read(nil)
			assert.NoError(t, err)
			assert.Equal(t, 0, n)
		})

		t.Run("use nil chan", func(t *testing.T) {
			var ch chans.Through[int]
			n, err := ch.Read(nil)
			buf := make([]int, 2)
			n, err = ch.Read(buf)
			assert.NoError(t, err)
			assert.Equal(t, 0, n)
		})
	})

	t.Run("creation with make", func(t *testing.T) {
		ch := make(chans.Through[int])
		assert.NotNil(t, ch)
		ch.Close()
	})
	t.Run("NewThrough", func(t *testing.T) {
		ch := chans.NewThrough[int]()
		assert.NotNil(t, ch)
		assert.Equal(t, 0, cap(ch))
		ch.Close()
	})
	t.Run("NewBufThrough", func(t *testing.T) {
		ch := chans.NewBufThrough[int](42)
		assert.NotNil(t, ch)
		assert.Equal(t, 42, cap(ch))
		ch.Close()
	})
	t.Run("NewPrefillThrough", func(t *testing.T) {
		ch := chans.NewPrefillThrough(41, 42)
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
		ch.Close()
	})
}
