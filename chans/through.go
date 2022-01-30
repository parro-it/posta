package chans

import "io"

// Through is a chan thet implement
// io.Reader, io.Writer, io.Closer
// when T is `byte`. Otherwise, the type
// implements chans.Reader[T], chans.Writer[T], io.Closer
type Through[T any] chan T

// NewThrough return a newly unbuffered Through instance
func NewThrough[T any]() Through[T] {
	return make(Through[T])
}

// NewThrough return a newly buffered Through instance
// with a buffer size of `capacity`
func NewBufThrough[T any](capacity int) Through[T] {
	return make(Through[T], capacity)
}

// NewPrefillThrough return a newly buffered Through instance
// with a buffer size sufficient to cache `data` elements.
// The buffer is then prefilled with all `data` elements.
func NewPrefillThrough[T any](data ...T) Through[T] {
	ch := make(Through[T], len(data))
	for _, v := range data {
		ch <- v
	}
	return ch
}

func (t Through[T]) Read(p []T) (n int, err error) {
	for i := 0; i < len(p); i++ {
		select {
		case v, ok := <-t:
			if !ok {
				return i, io.EOF
			}
			p[i] = v
		default:
			return i, nil
		}
	}
	return len(p), nil
}

func (t Through[T]) Write(p []T) (n int, err error) {
	for _, v := range p {
		t <- v
		n++
	}

	return
}

func (t Through[T]) Close() error {
	close(t)
	return nil
}
