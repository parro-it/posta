package chans

//type ThroughFunc[TIn any, TOut any] func(in TIn, out TOut)

type throughOptions[T any] struct {
	capacity *int
	initdata []T
	r        <-chan T
	w        chan<- T
}

type ThroughOptionsFn[T any] func(*throughOptions[T])

/*
func WithFunc[TIn any, TOut any](f ThroughFunc[TIn, TOut]) ThroughOptionsFn[TIn] {
	return func(o *throughOptions[TIn]) {
		//o.initdata = initdata
	}
}*/
func WithInitData[T any](initdata ...T) ThroughOptionsFn[T] {
	return func(o *throughOptions[T]) {
		o.initdata = initdata
	}
}

func WithCap[T any](capacity int) ThroughOptionsFn[T] {
	return func(o *throughOptions[T]) {
		o.capacity = &capacity
	}
}

// Through is a chan that implement
// io.Reader, io.Writer, io.Closer
// when T is `byte`. Otherwise, the type
// implements chans.Reader[T], chans.Writer[T], io.Closer
type Through[T any] struct {
	Reader[T]
	WriteCloser[T]
}

// NewThrough return a newly Through instance
func NewThrough[T any](options ...ThroughOptionsFn[T]) *Through[T] {
	var o throughOptions[T]

	for _, opfn := range options {
		opfn(&o)
	}
	var capacity int
	if o.capacity == nil {
		if o.initdata != nil {
			capacity = len(o.initdata)
		}
	} else {
		capacity = *o.capacity
		if o.initdata != nil && len(o.initdata) > capacity {
			panic("NewThrough: invalid input, capacity not sufficient to store initial data.")
		}
	}

	ch := make(chan T, capacity)
	if o.initdata != nil {
		for _, v := range o.initdata {
			ch <- v
		}
	}

	return &Through[T]{
		ChanReader[T](ch),
		ChanWriter[T](ch),
	}
}
