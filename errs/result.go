package errs

type Result[T any] struct {
	Res chan T
	Err error
}
