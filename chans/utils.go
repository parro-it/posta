package chans

import "io"

func ReadAll[T any](r Reader[T]) ([]T, error) {
	b := make([]T, 0, 512)
	for {
		if len(b) == cap(b) {
			var empty T
			// Add more capacity (let append pick how much).
			b = append(b, empty)[:len(b)] // slicing is required to unallocate the 'empty' element added
		}
		// read a chunk from current lenght
		// upto capacity
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n] // update b to include all elements just read above
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return b, err
		}
	}
}
