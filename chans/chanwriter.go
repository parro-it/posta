package chans

// ChanWriter is a chan that implement
// io.Writer and io.Closer
// when T is `byte`. Otherwise, the type
// implements chans.Writer[T] and io.Closer
type ChanWriter[T any] chan<- T

/*
 NewChanWriter return a new ChanWriter instance
 created according to specified options.

 Options can be any of:
   * FromChan - specifies a channel to use as the internal channel.
     If not specified, a new one is created.

   * WithCapacity - specifies a capacity for the channel. This
     option has no effects if a FromChan option is specified.
*/
func NewChanWriter[T any](options ...ChanOptionsFn[T]) (ChanWriter[T], <-chan T) {
	var o chanOptions[T]

	for _, opfn := range options {
		opfn(&o)
	}

	var ch chan T
	if o.source != nil {
		ch = (interface{})(o.source).(chan T)
	} else {
		var capacity int
		if o.capacity != nil {
			capacity = *o.capacity
		}
		ch = make(chan T, capacity)
	}
	return ChanWriter[T](ch), ch
}

func (t ChanWriter[T]) Write(p []T) (n int, err error) {
	for _, v := range p {
		t <- v
		n++
	}

	return
}

func (t ChanWriter[T]) Close() error {
	close(t)
	return nil
}
