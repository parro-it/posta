package actions2

// Action is an interface for actions
// send through the system
type Action any

// Queue send objects read from
// a channel to a set of registered listeners.
// New listeners could be created using
// ListenFor methods.
type Queue chan Action

func ListenFor[T any](q Queue) chan T {
	r := make(chan T)
	q <- addListener{l: listenerOf[T](r)}
	return r
}

func ListenFor2[T1 any, T2 any](q Queue) chan Action {
	r := make(chan Action)
	q <- addListener{l: listenerOf2[T1, T2](r)}
	return r
}

func ListenFor3[T1 any, T2 any, T3 any](q Queue) chan Action {
	r := make(chan Action)
	q <- addListener{l: listenerOf3[T1, T2, T3](r)}
	return r
}

func (q Queue) Unlisten(l any) {
	q <- removeListener{l: l}
}

// Start the gouroutine that
// continuously read objects
// from channel and send them
// to every registered listener.
func (q Queue) Start() Queue {
	if q == nil {
		q = make(Queue)
	}
	go q.start()
	return q
}

func (q Queue) start() {
	var listeners []listener

	for v := range q {
		switch action := v.(type) {
		case addListener:
			listeners = append(listeners, action.l)
		case removeListener:
			for i, l := range listeners {
				if l.Equal(action.l) {
					listeners = append(listeners[0:i], listeners[i+1:]...)
					l.Close()
					break
				}
			}
		case closeReq:
			for _, l := range listeners {
				l.Close()
			}
			break
		default:
			// forward the item to every reader
			for _, r := range listeners {
				r.Post(action)
			}
		}

	}
}

// an empty type to signal the goroutine
// to exit
type closeReq struct{}

// Close stops the gouroutine started
// by Start method. Also close
// any listener currently registered
func (q Queue) Close() {
	q <- closeReq{}
}

// an internal interface needed
// to abstract away the fact that
// listeners could be chan interface{}
// or chan T.
// This interface is implemented by these structs:
// - singleListener, which represents
// a listener who listens to a single object type.
// - multiListener, which represents
// a listener who listens to multiple types.
type listener interface {
	// Send an item to this listener.
	// The listener itself decides if the
	// item is to be send or discarded, according
	// for its requested types.
	Post(a Action)
	// Close the listener.
	Close()
	Equal(a any) bool
}
