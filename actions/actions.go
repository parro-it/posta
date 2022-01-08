package actions

type Action interface {
	Type() ActionType
}

var ch = make(chan Action)
var newReaders = make(chan Listener)

type ActionType int

type Listener interface {
	Post(a Action)
	Match(a Action) bool
	Close()
}

type listener struct {
	ch    chan Action
	types []ActionType
}

func (l listener) Post(a Action) {
	l.ch <- a
}
func (l listener) Close() {
	close(l.ch)
}
func (l listener) Match(a Action) bool {
	return contains(l.types, a.Type())

}
func Post(a Action) {
	ch <- a
}

func Listen(types ...ActionType) chan Action {
	r := make(chan Action)
	newReaders <- listener{
		ch:    r,
		types: types,
	}
	return r
}

type listenerOf[T any] struct {
	ch chan T
}

func (l listenerOf[T]) Post(a Action) {
	l.ch <- a.(T)
}
func (l listenerOf[T]) Close() {
	close(l.ch)
}
func (l listenerOf[T]) Match(a Action) bool {
	switch a.(type) {
	case T:
		return true
	default:
		return false
	}
}

func ListenOf[T any]() chan T {
	r := make(chan T)
	newReaders <- listenerOf[T]{
		ch: r,
	}
	return r
}

func contains[T comparable](s []T, e T) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Start() {
	var readers []Listener
	defer func() {
		for _, r := range readers {
			r.Close()
		}
		close(newReaders)
	}()

	for {
		select {
		case r := <-newReaders:
			readers = append(readers, r)
		case v := <-ch:
			if v == nil {
				continue
			}
			for _, r := range readers {
				if r.Match(v) {
					r.Post(v)
				}
			}
		}
	}
}
