package chans

import "sync"

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
	inputs sync.WaitGroup
}

func (q *SimpleMux[T]) Close() {
	q.inputs.Wait()
	close(q.Output)
}

func (q *SimpleMux[T]) AddInputFrom(in chan T) {
	if q.Output == nil {
		q.Output = make(chan T)
	}
	q.inputs.Add(1)
	go func() {
		for v := range in {
			q.Output <- v
		}
		q.inputs.Done()
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
