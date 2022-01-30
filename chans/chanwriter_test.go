package chans_test

import (
	"io"
	"testing"

	"github.com/parro-it/posta/chans"
	"github.com/stretchr/testify/assert"
)

func TestWriter(t *testing.T) {
	type empty struct{}
	w, _ := chans.NewChanWriter[empty]()
	var _ io.Closer = w
	var _ chans.Writer[empty] = w
	t.Run("Write", func(t *testing.T) {
		t.Run("2 values", func(t *testing.T) {
			ch, r := chans.NewChanWriter(chans.WithCapacity[int](2))
			n, err := ch.Write([]int{41, 42})
			assert.Equal(t, 2, n)
			assert.NoError(t, err)
			assert.Equal(t, 41, <-r)
			assert.Equal(t, 42, <-r)
			ch.Close()
			_, ok := <-r
			assert.False(t, ok)
		})
	})
}
