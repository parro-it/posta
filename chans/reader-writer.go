package chans

// Reader is the analogous of io.Reader, but for types other than `byte`.
// T is the type parameter that represent the type of data readed.
//
// Read reads up to len(p) T instances into p. It returns the number of T instances
// read (0 <= n <= len(p)) and any error encountered. Even if Read
// returns n < len(p), it may use all of p as scratch space during the call.
// If some data is available but not len(p) T instances, Read conventionally
// returns what is available instead of waiting for more.
//
// When Read encounters an error or end-of-file condition after
// successfully reading n > 0 T instances, it returns the number of
// instances read. It may return the (non-nil) error from the same call
// or return the error (and n == 0) from a subsequent call.
// An instance of this general case is that a Reader returning
// a non-zero number of instances at the end of the input stream may
// return either err == EOF or err == nil. The next Read should
// return 0, EOF.
//
// Callers should always process the n > 0 instances returned before
// considering the error err. Doing so correctly handles I/O errors
// that happen after reading some instances and also both of the
// allowed EOF behaviors.
//
// Implementations of Read are discouraged from returning a
// zero count with a nil error, except when len(p) == 0.
// Callers should treat a return of 0 and nil as indicating that
// nothing happened; in particular it does not indicate EOF.
//
// Implementations must not retain p.
type Reader[T any] interface {
	Read(p []T) (n int, err error)
}

// Writer is the interface that wraps the basic Write method.
//
// Write writes len(p) T instances from p to the underlying data stream.
// It returns the number of instances written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// Write must return a non-nil error if it returns n < len(p).
// Write must not modify the slice data, even temporarily.
//
// Implementations must not retain p.
type Writer[T any] interface {
	Write(p []T) (n int, err error)
}
