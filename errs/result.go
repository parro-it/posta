package errs

type Result[T any] struct {
	Res chan T
	Err error
}

func NewResult[T any]() Result[T] {
	return Result[T]{
		Res: make(chan T),
	}
}

func (r *Result[_]) End(err error) {

}

func (r *Result[T]) Try(val T, err error) T {
	if err != nil {
		var empty T
		r.Err = err

		return empty
	}
	return val
}
