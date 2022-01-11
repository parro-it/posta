package plex

// Demux reads data from an input
// channel and ditributes it
// to a set of output channels.
//
// New output channels could be added
// after the Demux creation using AddOut
// methods. Newly added channels will only
// receives data sent to the input chan after
// they was added.
type Demux chan any

// AddOut creates a new chan T
// and add it to the set of output
// channels.
// The new output will only receive
// sent messages of type T
func AddOut[T any](q Demux) chan T {
	r := make(chan T)
	q <- addOut{out: outOf[T](r)}
	return r
}

// AddOut2 creates a new chan any
// and add it to the set of output
// channels.
// The new output will only receive
// sent messages of type T1 or T2
func AddOut2[T1 any, T2 any](q Demux) chan any {
	r := make(chan any)
	q <- addOut{out: outOf2[T1, T2](r)}
	return r
}

// AddOut3 creates a new chan any
// and add it to the set of output
// channels.
// The new output will only receive
// sent messages of type T1 or T2 or T3
func AddOut3[T1 any, T2 any, T3 any](q Demux) chan any {
	r := make(chan any)
	q <- addOut{out: outOf3[T1, T2, T3](r)}
	return r
}

// Remove an output channel from
// the set of registered ones.
// The output channel is also closed
// after it has been removed.
func (q Demux) RemoveOut(l any) {
	q <- removeOut{out: l}
}

// Start the internal gouroutine that
// continuously read data from input
// channel and distribute them to every
// registered output.
//
// The gouroutine also listen for
// 3 unexported types in the input data that
// instruct it to do specific actions. These 3
// types are addOut, removeOut and closeReq.
// The first two types cause the addition and removal
// of outputs, closeReq completely closes the
// Demux and terminates the gouroutin itself.
func (q Demux) Start() Demux {
	if q == nil {
		q = make(Demux)
	}
	go q.start()
	return q
}

func (q Demux) start() {
	var outputs []output

	for data := range q {
		switch msg := data.(type) {
		case addOut:
			outputs = append(outputs, msg.out)
		case removeOut:
			for i, out := range outputs {
				if out.Equal(msg.out) {
					outputs = append(outputs[0:i], outputs[i+1:]...)
					out.Close()
					break
				}
			}
		case closeReq:
			for _, out := range outputs {
				out.Close()
			}
			break
		default:
			// forward the item to every reader
			for _, r := range outputs {
				r.Post(msg)
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
func (q Demux) Close() {
	q <- closeReq{}
}

// an internal interface needed
// to abstract away the fact that
// outputs channels could have type
// `chan interface{}` when listening
// for multiple types or `chan T`` when listening
// single types.
// This interface is implemented by these structs:
// - singleType, which represents
// a output who listens to a single type.
// - multiTypes, which represents
// a output who listens to multiple types.
type output interface {
	// Send an item to this output channel.
	// The implementing type must decides if the
	// item is to be send or discarded, according
	// for its requested types.
	Post(a any)
	// Close the channel.
	Close()
	Equal(a any) bool
}
