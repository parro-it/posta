package chans

import "io"

// ChanReader is a chan that implement
// io.Reader when T is `byte`. Otherwise, the type
// implements chans.Reader[T]
type ChanReader[T any] <-chan T

type newChanReaderOptions[T any] struct {
	capacity *int
	initdata []T
	source   <-chan T
}

type ChanReaderOptionsFn[T any] func(*newChanReaderOptions[T])

func PrefilledWith[T any](initdata ...T) ChanReaderOptionsFn[T] {
	return func(o *newChanReaderOptions[T]) {
		o.initdata = initdata
	}
}

func FromChan[T any](source <-chan T) ChanReaderOptionsFn[T] {
	return func(o *newChanReaderOptions[T]) {
		o.source = source
	}
}

func WithCapacity[T any](capacity int) ChanReaderOptionsFn[T] {
	return func(o *newChanReaderOptions[T]) {
		o.capacity = &capacity
	}
}

/*
 NewChanReader return a new ChanReader instance
 created according to specified options.

 Options can be any of:
   * PrefilledWith - specifies an initial set of T instances
     that are sent to the internal channel. If a WithCapacity
     option is not specified, a capacity is used equal to len
	 of data, to avoid blocking. If a WithCapacity is specified,
	 and it is smaller than PrefilledWith data len, the
     funcion panics.

   * FromChan - specifies a channel to use as the internal channel.
     If not specified, a new one is created.

   * WithCapacity - specifies a capacity for the channel. This
     option has no effects if a FromChan option is specified.
*/
func NewChanReader[T any](options ...ChanReaderOptionsFn[T]) (ChanReader[T], chan<- T) {
	var o newChanReaderOptions[T]

	for _, opfn := range options {
		opfn(&o)
	}

	var ch chan T
	if o.source != nil {
		ch = (interface{})(o.source).(chan T)
	} else {
		var capacity int
		if o.capacity == nil {
			if o.initdata != nil {
				capacity = len(o.initdata)
			}
		} else {
			capacity = *o.capacity
			if o.initdata != nil && len(o.initdata) > capacity {
				panic("NewChanReader: invalid input, capacity not sufficient to store initial data.")
			}
		}
		ch = make(chan T, capacity)
	}

	if o.initdata != nil {
		for _, v := range o.initdata {
			ch <- v
		}
	}

	return ChanReader[T](ch), ch
}

func (t ChanReader[T]) Read(p []T) (n int, err error) {
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
