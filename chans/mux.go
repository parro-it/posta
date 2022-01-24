package chans

type Out[T any] struct {
	Value T
	Idx   int
}

type Mux[T any] struct {
	Output     chan Out[T]
	inputCount int
}

type SimpleMux[T any] struct {
	Output chan T
}

func (q *SimpleMux[T]) AddInputFrom(in chan T) {
	if q.Output == nil {
		q.Output = make(chan T)
	}
	go func() {
		for v := range in {
			q.Output <- v
		}
	}()
}
func (q *Mux[T]) AddInputFrom(in chan T) {
	if q.Output == nil {
		q.Output = make(chan Out[T])
	}
	idx := q.inputCount
	q.inputCount++
	go func() {
		for v := range in {
			q.Output <- Out[T]{Value: v, Idx: idx}
		}
	}()
}
