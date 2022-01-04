package actions

type Action interface {
	Type() ActionType
}

var ch = make(chan Action)
var newReaders = make(chan Listener)

type ActionType int

type Listener struct {
	ch    chan Action
	types []ActionType
}

func Post(a Action) {
	ch <- a
}

func Listen(types ...ActionType) chan Action {
	r := make(chan Action)
	newReaders <- Listener{
		ch:    r,
		types: types,
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
			close(r.ch)
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
				if contains(r.types, v.Type()) {
					r.ch <- v
				}
			}
		}
	}
}
