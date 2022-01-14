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
//
// Demux has an internal goroutine
// that do the demultiplexing that
// must be started by calling the Start
// methods.
//
// An existing channel could be used
// by setting Input field before calling
// Start method. Start method create a chan
// if the field is nil.
//
// Values could be sent to the Input field
// channel, after Start method is called.
//
// It's an error to set Input field after
// Start method is called.
// It's an error to send to Input field channel before
// Start method is called.
type Demux[T any] struct {
	Input    chan T
	commands chan any
}

// AddOut creates a new chan T
// and add it to the set of output
// channels.
// The new output will only receive
// sent messages of type T
func AddOut[T any, TInput any](q Demux[TInput]) chan T {
	r := make(chan T)
	q.commands <- addOut{out: outOf[T](r)}
	return r
}

// AddOut2 creates a new chan any
// and add it to the set of output
// channels.
// The new output will only receive
// sent messages of type T1 or T2
func AddOut2[T1 any, T2 any, TInput any](q Demux[TInput]) chan any {
	r := make(chan any)
	q.commands <- addOut{out: outOf2[T1, T2](r)}
	return r
}

// AddOut3 creates a new chan any
// and add it to the set of output
// channels.
// The new output will only receive
// sent messages of type T1 or T2 or T3
func AddOut3[T1 any, T2 any, T3 any, TInput any](q Demux[TInput]) chan any {
	r := make(chan any)
	q.commands <- addOut{out: outOf3[T1, T2, T3](r)}
	return r
}

// Remove an output channel from
// the set of registered ones.
// The output channel is also closed
// after it has been removed.
func (q Demux[_]) RemoveOut(l any) {
	q.commands <- removeOut{out: l}
}

// Start the internal gouroutine that
// continuously read data from input
// channel and distribute them to every
// registered output.
//
// The gouroutine also read from internal command
// chan and execute either an addOut or removeOut command.
// The commands causes the addition and removal
// of outputs.
func (q *Demux[T]) Start() {
	q.commands = make(chan any)
	if q.Input == nil {
		q.Input = make(chan T, 1024)
	}

	go func() {
		var outputs []output

		defer func() {
			// when the gouroutine terminates,
			// the commands chan and all output
			// chan are closed
			close(q.commands)
			for _, out := range outputs {
				out.Close()
			}
		}()

		for {
			select {
			case data, isOpen := <-q.Input:
				if !isOpen {
					// the gouroutine terminates when
					// Input chan is closed
					return
				}
				//fmt.Printf("%s %v\n", reflect.TypeOf(data).Name(), data)

				// forward the item to every reader
				for _, r := range outputs {
					r.Post(data)
				}
			case cmd := <-q.commands:
				//fmt.Printf("%s\n", cmd)
				switch cmd := cmd.(type) {
				case addOut:
					// register a new output channel
					outputs = append(outputs, cmd.out)
				case removeOut:
					// remove a registered channel
					for i, out := range outputs {
						if out.Equal(cmd.out) {
							outputs = append(outputs[0:i], outputs[i+1:]...)
							out.Close()
							break
						}
					}

				}
			}
		}
	}()

}

// Close stops the gouroutine started
// by Start method. Also close
// any listener currently registered
func (q Demux[_]) Close() {
	close(q.Input)
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
