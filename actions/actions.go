package actions

type Action any

var ch = make(chan Action)
var newReaders = make(chan chan Action)

func Post(a Action) {
	ch <- a
}

func AddReader() chan Action {
	r := make(chan Action)
	newReaders <- r
	return r
}

func Start() {
	var readers []chan Action
	defer func() {
		for _, r := range readers {
			close(r)
		}
		close(newReaders)
	}()

	for {
		select {
		case r := <-newReaders:
			if r == nil {
				close(ch)
				return
			}
			readers = append(readers, r)
		case v := <-ch:
			if v == nil {
				continue
			}
			for _, r := range readers {
				r <- v
			}
		}
	}
}
