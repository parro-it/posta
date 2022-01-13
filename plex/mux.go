package plex

type Out[T any] struct {
	Value T
	Idx   int
}

type Mux[T any] struct {
	Output     chan Out[T]
	inputCount int
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
